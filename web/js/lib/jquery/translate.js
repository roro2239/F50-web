(function ($) {
    $.fn.translate = function () {
    	var $this = $(this);
    	$this.each(function(){
    		var item = $(this);
    		var trans = item.attr("data-trans");
    		if(!!trans){
            	translateElement(this, trans);
    		}

            var placeholder = item.attr("data-placeholder");
            if(!!placeholder){
                placeholderElement(this, placeholder);
            }
    	});

        $this.find("*[data-tooltip-trans]").each(function() {
            var item = $(this);
            var trans = item.attr("data-tooltip-trans");
            var transData = item.attr("data-tooltip-trans-data");
            var word;
            if(transData == undefined) {
                word = $.i18n.prop(trans);
            } else {
                word = $.i18n.prop.apply(null, [trans].concat(transData.split(",")));
            }

            // item.attr("title", word);
            item.attr("data-bs-original-title", word);
        });

    	$this.find("*[data-trans], *[data-placeholder]").each(function () {
            var self = $(this);
    		if(self.attr("id") == 'chosen-search-field-input'){
    			var val = $("#chosenUserSelect").val();
    			if(val && val.length > 0){
    				return;
    			}
    		}
            var trans = self.attr("data-trans");
            var transData = self.attr("data-trans-data");
            if (!!trans) {
            	translateElement(this, trans, transData);
            }

            var placeholder = self.attr("data-placeholder");
            if(!!placeholder){
                placeholderElement(this, placeholder);
            }
        });

    	//翻译国家码
        $('*[data-transid]', $this).each(function () {
        	var ele = $(this);
            var transid = ele.attr('data-transid');
            if(ele.attr("name") == "channel" || ele.attr("name") == "channels_5g" ||ele.attr("name") == "guestChannel"){
            	ele.find('option').each(function () {
            		var item = $(this);
            		if (item.val() != 0) {
            			var val = item.val().split("_");
						if(val[0] == '3664100140' || val[0] =='3664100112132140' || val[0] == '3648'){
							item.html( $.i18n.prop(transid + '_' + val[0]) );
						} else {
							item.html( val[1] + "MHz " + $.i18n.prop(transid + '_' + val[0]) );
						}
            		} else {
            			item.html( $.i18n.prop(transid + '_0') );
            		}
            	});
            }else{
            	ele.find('option').each(function () {
            		$(this).html($.i18n.prop(transid + '_' + $(this).attr('value')));
            	});
            }
        });

        $('*[transId]', $this).each(function () {
        	var ele = $(this);
            var transid = ele.attr('transId');
            if(ele.attr("id").indexOf('access_week_txt_') != -1){
            	var transArray = transid.split(',');
				var textTrans = "";
				for(var i = 0;i<transArray.length;i++){
					if(textTrans==''){
						textTrans = textTrans + $.i18n.prop(transArray[i]);
					}else{						
						textTrans = textTrans + ',' + $.i18n.prop(transArray[i]);
					}		
				}
				$('#'+ele.attr("id")).html(textTrans);
            }
        });

        function translateElement(ele, trans, transData){
            var word;
            if(transData == undefined) {
                word = $.i18n.prop(trans);
            } else {
                word = $.i18n.prop.apply(null, [trans].concat(transData.split(",")));
            }
            var nodeName = ele.nodeName.toUpperCase();
            if (nodeName == 'INPUT' || nodeName == 'SELECT' || nodeName == 'TEXTAREA') {
                $(ele).val(word);
            } else if (nodeName == 'BUTTON') {
                $(ele).text(word);
            } else {
                $(ele).html(word);
            }
        }

        function placeholderElement(ele, trans){
            var word = $.i18n.prop(trans);
            var nodeName = ele.nodeName.toUpperCase();
            if (nodeName == 'INPUT') {
                $(ele).attr('placeholder', word);
            }
        }

        $('form div.row', $this).each(function () {
            var $row = $(this);
            if ($row.has('.required').length > 0) {
                $("label:first-child", $row).append("<i class='colorRed'>&nbsp;*</i>");
            } else {
                $("label:first-child", $row).append("<i class='colorRed' style='visibility: hidden;'>&nbsp;*</i>");
            }
        });

        return $this;
    };
})(jQuery);
