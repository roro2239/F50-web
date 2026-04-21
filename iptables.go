package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

//改自 github.com/pretty66/iptables-web
var iptabser Iptableser

type Protocol byte

const (
	ProtocolIPv4 Protocol = iota
	ProtocolIPv6
)

type IptablesV4CMD struct {
	binary        string
	saveBinary    string
	restoreBinary string
	protocol      Protocol
	exec          exec.Cmd
}

type option func(*IptablesV4CMD)

func NewIPV4(opt ...option) (*IptablesV4CMD, error) {
	ipc := &IptablesV4CMD{}
	for _, fn := range opt {
		fn(ipc)
	}
	if ipc.protocol == ProtocolIPv6 {
		return nil, errors.New("IPv6 is not supported temporarily")
	}
	if len(ipc.binary) == 0 {
		ipc.binary = "iptables"
	}
	if len(ipc.saveBinary) == 0 {
		ipc.saveBinary = "iptables-save"
	}
	if len(ipc.restoreBinary) == 0 {
		ipc.restoreBinary = "iptables-restore"
	}
	return ipc, nil
}

func WithProtocol(protocol Protocol) option {
	return func(ic *IptablesV4CMD) {
		ic.protocol = protocol
	}
}

func WithBinary(cmd string) option {
	return func(ic *IptablesV4CMD) {
		ic.binary = cmd
	}
}

func WithSaveBinary(cmd string) option {
	return func(ic *IptablesV4CMD) {
		ic.saveBinary = cmd
	}
}

func WithRestoreBinary(cmd string) option {
	return func(ic *IptablesV4CMD) {
		ic.restoreBinary = cmd
	}
}

func (i *IptablesV4CMD) Version() (string, error) {
	return i.iptables("--version")
}

func (i *IptablesV4CMD) ListRule(table, chain string) (map[string][]TableList, error) {
	if len(table) == 0 {
		table = "filter"
	}
	var str string
	var err error
	if len(chain) == 0 {
		str, err = i.iptables("-t", table, "-nvL", "--line-numbers")
	} else {
		str, err = i.iptables("-t", table, "-L", chain, "-nv", "--line-numbers")
	}

	if err != nil {
		return nil, err
	}

	tl := map[string][]TableList{}
	tl["system"] = make([]TableList, 0)
	tl["custom"] = make([]TableList, 0)

	chains := iSplitAndTrimSpace(str, "\n\n")
	for k := range chains {
		column := []Column{}
		chainList := iSplitAndTrimSpace(chains[k], "\n")
		if len(chainList) == 0 {
			continue
		}
		if len(chainList) > 2 {
			column, err = parseColumn(chainList[2:])
			if err != nil {
				log.Println(err)
				continue
			}
		}

		stitle, err := parseSystemTitle(chainList[0])
		if err == nil {
			tl["system"] = append(tl["system"], SystemTable{
				SystemTitle: stitle,
				Column:      column,
			})
		} else {
			ctitle, err := parseCustomTitle(chainList[0])
			if err != nil {
				log.Println(err)
				continue
			}
			tl["custom"] = append(tl["custom"], CustomTable{
				CustomTitle: ctitle,
				Column:      column,
			})
		}
	}
	return tl, nil
}

func (i *IptablesV4CMD) FlushRule(table, chain string) error {
	var err error
	if len(table) == 0 && len(chain) == 0 {
		_, err = i.iptables("-t", "raw", "-F")
		_, err = i.iptables("-t", "mangle", "-F")
		_, err = i.iptables("-t", "nat", "-F")
		_, err = i.iptables("-t", "filter", "-F")
		return err
	}

	if len(table) == 0 {
		table = "filter"
	}
	if len(chain) == 0 {
		_, err = i.iptables("-t", table, "-F")
	} else {
		_, err = i.iptables("-t", table, "-F", chain)
	}

	return err
}

func (i *IptablesV4CMD) FlushMetrics(table, chain, id string) error {
	var err error
	if len(id) > 0 {
		if len(table) == 0 || len(chain) == 0 {
			return fmt.Errorf("FlushMetrics args error. table:%s chain:%s id:%s", table, chain, id)
		}
		_, err = i.iptables("-t", table, "-Z", chain, id)
		return err
	}

	if len(table) == 0 && len(chain) == 0 {
		_, err = i.iptables("-t", "raw", "-Z")
		_, err = i.iptables("-t", "mangle", "-Z")
		_, err = i.iptables("-t", "nat", "-Z")
		_, err = i.iptables("-t", "filter", "-Z")
		return err
	}

	if len(table) == 0 {
		table = "filter"
	}
	if len(chain) == 0 {
		_, err = i.iptables("-t", table, "-Z")
	} else {
		_, err = i.iptables("-t", table, "-Z", chain)
	}

	return err
}

func (i *IptablesV4CMD) DeleteRule(table, chain, id string) error {
	if len(table) == 0 || len(chain) == 0 || len(id) == 0 {
		return fmt.Errorf("DeleteRule args error. table:%s chain:%s id:%s", table, chain, id)
	}
	_, err := i.iptables("-t", table, "-D", chain, id)
	return err
}

func (i *IptablesV4CMD) ListExec(table, chain string) (string, error) {
	var str string
	var err error
	if len(chain) == 0 {
		str, err = i.iptablesSave("-t", table)
	} else {
		// chain不用去除空格，显示引用命令
		str, err = i.iptablesSave("-t", table, "|", "grep", chain)
	}
	if err != nil {
		log.Println("ListExec:", err)
	}
	return str, err
}

func (i *IptablesV4CMD) Exec(param ...string) (string, error) {
	var args []string
	for k := range param {
		param[k] = strings.TrimSpace(param[k])
		if len(param[k]) == 0 {
			continue
		}
		args = append(args, param[k])
	}
	return i.iptables(args...)
}

func (i *IptablesV4CMD) GetRuleInfo(table, chain, id string) (string, error) {
	if len(table) == 0 || len(chain) == 0 || len(id) == 0 {
		return "", fmt.Errorf("GetRuleInfo args error. table:%s chain:%s id:%s", table, chain, id)
	}
	// s, err := i.iptablesSave(fmt.Sprintf("-t %s | grep %s", table, " "+chain+" "))
	s, err := i.iptablesSave(fmt.Sprintf("-t %s | grep ' %s '", table, chain))
	if err != nil {
		return "", err
	}
	list := iSplitAndTrimSpace(s, "\n")
	idint, _ := strconv.Atoi(id)
	if len(list) < idint {
		return "", fmt.Errorf("GetRuleInfo rule not found. table:%s chain:%s id:%s", table, chain, id)
	}
	return list[idint-1], nil
}

func (i *IptablesV4CMD) FlushEmptyCustomChain() error {
	_, err := i.iptables("-t", "raw", "-X")
	_, err = i.iptables("-t", "mangle", "-X")
	_, err = i.iptables("-t", "nat", "-X")
	_, err = i.iptables("-t", "filter", "-X")
	return err
}

func (i *IptablesV4CMD) Export(table, chain string) (string, error) {
	var args []string
	if len(table) > 0 {
		args = append(args, table)
	}
	if len(chain) > 0 {
		args = append(args, chain)
	}
	return i.iptablesSave(args...)
}

func (i *IptablesV4CMD) Import(rule string) error {
	if len(rule) == 0 {
		return nil
	}
	fileName := "/data/local/tmp/iptable.rule"
	err := os.WriteFile(fileName, []byte(rule), fs.ModePerm)
	if err != nil {
		return fmt.Errorf("Import rule error. err:%v", err)
	}
	defer os.Remove(fileName)
	_, err = i.iptablesRestore(fileName)
	return err
}

func (i *IptablesV4CMD) iptables(args ...string) (string, error) {
	return rootShellLog(i.binary + " " + strings.Join(args, " "))
}

func (i *IptablesV4CMD) iptablesSave(args ...string) (string, error) {
	return rootShellLog(fmt.Sprintf("%s %s", i.saveBinary, strings.Join(args, " ")))
}

func (i *IptablesV4CMD) iptablesRestore(fileName string) (string, error) {
	return rootShellLog(fmt.Sprintf("%s < %s", i.restoreBinary, fileName))
}

type SystemTitle struct {
	Chain   string `json:"chain"`   // PREROUTING、INPUT、OUTPUT、POSTROUTING
	Policy  string `json:"policy"`  // 默认策略：ACCEPT、DROP
	Packets string `json:"packets"` // 包数量
	Bytes   string `json:"bytes"`   // 字节数
}

type CustomTitle struct {
	Chain      string `json:"chain"`      // PREROUTING、INPUT、OUTPUT、POSTROUTING
	References string `json:"references"` // 引用数量
}

type Column struct {
	Num         string `json:"num"`
	Pkts        string `json:"pkts"`
	Bytes       string `json:"bytes"`
	Target      string `json:"target"`
	Prot        string `json:"prot"`
	Opt         string `json:"opt"`
	In          string `json:"in"`
	Out         string `json:"out"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Action      string `json:"action"`
}

type SystemTable struct {
	SystemTitle `json:"title"`
	Column      []Column `json:"list"`
}

func (st SystemTable) String() string {
	return string(iJSONEncoding(st))
}

type CustomTable struct {
	CustomTitle `json:"title"`
	Column      []Column `json:"list"`
}

func (st CustomTable) String() string {
	return string(iJSONEncoding(st))
}

type TableList interface {
	String() string
}

var (
	systemTitleRegex *regexp.Regexp
	customTitleRegex *regexp.Regexp
	columnRegex      *regexp.Regexp
)

func init() {
	var err error
	systemTitleRegex, err = regexp.Compile(`Chain (.+) \(policy (.+) (.+) packets, (.+) bytes\)`)
	if err != nil {
		panic(err)
	}
	customTitleRegex, err = regexp.Compile(`Chain (.+) \((\d+) references\)`)
	if err != nil {
		panic(err)
	}
	columnRegex, err = regexp.Compile(`(\d+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+([0-9\.\/]+)\s*(.*)`)
	if err != nil {
		panic(err)
	}
}

func parseSystemTitle(ts string) (out SystemTitle, err error) {
	res := systemTitleRegex.FindStringSubmatch(ts)
	if len(res) != 5 {
		err = fmt.Errorf("parse system table title error:%d => %v", len(res), res)
		return
	}
	out.Chain = res[1]
	out.Policy = res[2]
	out.Packets = res[3]
	out.Bytes = res[4]
	return
}

func parseCustomTitle(ts string) (out CustomTitle, err error) {
	res := customTitleRegex.FindStringSubmatch(ts)
	if len(res) != 3 {
		err = fmt.Errorf("parse custom table title error:%d => %v", len(res), res)
		return
	}
	out.Chain = res[1]
	out.References = res[2]
	return
}

func parseColumn(cs []string) ([]Column, error) {
	out := []Column{}
	for k := range cs {
		if len(cs[k]) == 0 {
			continue
		}
		rule := columnRegex.FindStringSubmatch(cs[k])
		if len(rule) < 12 {
			return nil, fmt.Errorf("parse column error:%d => %v, str:%s", len(rule), rule, cs[k])
		}
		rule = rule[1:]
		c := Column{
			Num:         rule[0],
			Pkts:        rule[1],
			Bytes:       rule[2],
			Target:      rule[3],
			Prot:        rule[4],
			Opt:         rule[5],
			In:          rule[6],
			Out:         rule[7],
			Source:      rule[8],
			Destination: rule[9],
			Action:      rule[10],
		}
		out = append(out, c)
	}
	return out, nil
}

type Iptableser interface {
	Version() (string, error)
	ListRule(table, chain string) (map[string][]TableList, error)
	FlushRule(table, chain string) error
	FlushMetrics(table, chain, id string) error
	DeleteRule(table, chain, id string) error
	ListExec(table, chain string) (string, error)
	Exec(param ...string) (string, error)
	GetRuleInfo(table, chain, id string) (string, error)
	FlushEmptyCustomChain() error
	Export(table, chain string) (string, error)
	Import(rule string) error
}

func init() {
	var err error
	iptabser, err = NewIPV4()
	if err != nil {
		print(err)
		return
	}
}

func iptablesAPI(w http.ResponseWriter, req *http.Request) {
	if !isLogin(w, req) {
		json.NewEncoder(w).Encode(map[string]string{"loginfo": "no"})
		return
	}
	// 获取请求路径
	path := req.URL.Path
	path = strings.TrimPrefix(path, "/iptablesAPI/")
	switch path {
	case "version":
		v, err := iptabser.Version()
		iOutput(w, err, v)
	case "listRule":
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		l, err := iptabser.ListRule(table, chain)
		iOutput(w, err, l)
	case "listExec":
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		l, err := iptabser.ListExec(table, chain)
		iOutput(w, err, l)
	case "flushRule":
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		err := iptabser.FlushRule(table, chain)
		iOutput(w, err, nil)
	case "deleteRule":
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		id := req.FormValue("id")
		err := iptabser.DeleteRule(table, chain, id)
		iOutput(w, err, nil)
	case "flushMetrics":
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		id := req.FormValue("id")
		err := iptabser.FlushMetrics(table, chain, id)
		iOutput(w, err, nil)

	case "getRuleInfo":
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		id := req.FormValue("id")
		info, err := iptabser.GetRuleInfo(table, chain, id)
		iOutput(w, err, info)

	case "flushEmptyCustomChain":
		err := iptabser.FlushEmptyCustomChain()
		iOutput(w, err, nil)
	case "export":
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		rule, err := iptabser.Export(table, chain)
		iOutput(w, err, rule)

	case "import":
		rule := req.FormValue("rule")
		err := iptabser.Import(rule)
		iOutput(w, err, nil)

	case "exec":
		args := req.FormValue("args")
		if len(args) == 0 {
			iOutput(w, nil, nil)
			return
		}
		s := iSplitAndTrimSpace(args, " ")
		str, err := iptabser.Exec(s...)
		iOutput(w, err, str)

	}

}
func iJSONEncoding(data interface{}) []byte {
	b, _ := json.Marshal(data)
	return b
}

func iOutput(w http.ResponseWriter, err error, data interface{}) {
	var code int
	msg := "OK"
	if err != nil {
		code = 1
		msg = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
	out := map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": data,
	}
	_, _ = w.Write(iJSONEncoding(out))
}
func iSplitAndTrimSpace(s, sep string) []string {
	res := strings.Split(s, sep)
	for k := range res {
		res[k] = strings.TrimSpace(res[k])
	}
	return res
}
