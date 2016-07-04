// Selectize-Knockout bridge
// Original source : http://jsfiddle.net/xtranophilist/hDefE/5/

var inject_binding = function (allBindings, key, value){
	//https://github.com/knockout/knockout/pull/932#issuecomment-26547528
	return {
		has: function (bindingKey) {
			return (bindingKey == key) || allBindings.has(bindingKey);
		},
		get: function (bindingKey) {
			var binding = allBindings.get(bindingKey);
			if (bindingKey == key) {
				binding = binding ? [].concat(binding, value) : value;
			}
			return binding;
		}
	};	
}; // end inject_binding

ko.bindingHandlers.selectize = {
    init: function (element, valueAccessor, allBindingsAccessor, viewModel, bindingContext) {

        if (!allBindingsAccessor.has('optionsText'))
            allBindingsAccessor = inject_binding(allBindingsAccessor, 'optionsText', 'name');
        if (!allBindingsAccessor.has('optionsValue'))
            allBindingsAccessor = inject_binding(allBindingsAccessor, 'optionsValue', 'id');
        if (typeof allBindingsAccessor.get('optionsCaption') == 'undefined')
            allBindingsAccessor = inject_binding(allBindingsAccessor, 'optionsCaption', 'Choose...');

        ko.bindingHandlers.options.update(element, valueAccessor, allBindingsAccessor, viewModel, bindingContext);

        var options = {
            valueField: allBindingsAccessor.get('optionsValue'),
            labelField: allBindingsAccessor.get('optionsText'),
            searchField: allBindingsAccessor.get('optionsText')
        };

        var $select = $(element).selectize(options)[0].selectize;

        // Specific to application API integration
        if (typeof allBindingsAccessor.get('endpoint') != 'undefined') {
            options.load = function(query, callback) {
                if (!query.length) return callback();
                $.ApiGET = function(allBindingsAccessor.get('endpoint') + '/' + encodeURIComponent(query), function(res) {
                    callback(res.repositories.slice(0, 10));
                });
            };
        } else {
            $select.addItem(allBindingsAccessor.get('value')());
        }

        if (typeof init_selectize == 'function') {
            init_selectize($select);
        }

        valueAccessor().subscribe(function (new_value) {
            var new_obj = new_value[new_value.length - 1];
            $select.addOption(new_obj);
        });
    }, // end init
    update: function (element, valueAccessor, allBindingsAccessor) {
        if (allBindingsAccessor.has('object')) {
            var value_accessor = valueAccessor();
            var selected_obj = $.grep(value_accessor(), function (i) {
                return i.id == allBindingsAccessor.get('value')();
            })[0];

            if (selected_obj) {
                allBindingsAccessor.get('object')(selected_obj);
            }
        }
    } // end update
}; // end ko.bindingHandlers.selectize

