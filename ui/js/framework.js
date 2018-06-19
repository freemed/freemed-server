// framework.js: jQuery + Knockout + Bootstrap framework
// @jbuchbinder


////////// Global Variable and Settings //////////

var apiBase = "../api";
var currentPage = null;
var sessionId;
var startPage = 'main';
var pageParams = null;
var globalTimeout = 5000; // ms
var sessionExpiry = 600; // seconds


////////// Authentication Functions //////////

var sessionAuth = function(xhr) {
	xhr.setRequestHeader('Authorization', 'Bearer ' + sessionId);
};

function storeSessionId(sessionId) {
	$.sessionStorage.setItem('sessionId', JSON.stringify({
		'sessionId': sessionId,
		'expiry': new Date().value + (1000 * sessionExpiry)
	}));
}

function authenticated() {
	return sessionId != null;
} // end function authenticated

function login() {
	$.ajax({
		url: apiBase + "/../auth/login",
		method: "POST",
		contentType: "application/json",
		data: JSON.stringify({
			username: $('#login-username').val(),
			password: $('#login-password').val()
		}),
		error: function(x){
			console.log(JSON.stringify(x));
			toastr.error('Unable to login -- please try again.', 'Login', {timeOut: 5000});
			$('#login-password').val(''); // Clear password, for security purposes
			loginStateChange(false, null);
		},
		success: function(data){
			sessionId = data.token;
			storeSessionId(sessionId);
			loginStateChange(true, function() {
				$('#login-password').val(''); // Clear password, for security purposes
				console.log('cb: sessionId = ' + sessionId);
				if (currentPage == null || currentPage == 'login-splash') {
					loadPage('main');
				}
			});
		}
	});
} // end function login

function loginStateChange(loggedin, cb) {
	if (loggedin) {
		console.log('Login successful');
		$( 'LI.nav-authed' ).show();
		$( '#loginDialog' ).modal('hide');
		if (cb != null) {
			cb();
		}
	} else {
		console.log('Login failed');
		$( 'LI.nav-authed' ).hide();
		$( '#loginDialog' ).modal('show');
		$( '#login-username' ).focus();
	}
} // end function loginStateChange
 
function logout() {
	$.ajax({
		url: apiBase + "/../auth/logout",
		method: "DELETE",
		contentType: "application/json",
		beforeSend: sessionAuth,
		error: displayError,
		success: function(data) {
			$.sessionStorage.removeItem('sessionId');
			loadPage('login-splash');
			loginStateChange(false, null);
		}
	});
} // end function logout

function displayError( err ) {
	$( '#pageAlert' ).html( err );
	$( '#pageAlert' ).addClass( 'alert-error' );
	$( '#pageAlert' ).show( "slow", "fade", function() {
		setTimeout(function() {
			$( '#pageAlert' ).hide( "slow", "fade", function() {
				$( '#pageAlert' ).removeClass( 'alert-error' );
			});
		}, 3000);
	});
} // end function displayError

function loadPage( id ) {
	console.log('Loading page ' + id);

	var origId = id;
	var idx = id.indexOf('#');
	if (idx != -1) {
		// Pass parameters
		pageParams = id.substr(idx+1);
		id = id.substr(0, idx);
		console.log("loadPage(): pageParams = '" + pageParams + "', id = '" + id + "'");
	}

	var ts = new Date().getTime();

	if (id != 'login-splash') {
		console.log('Removing bindings from #mainFrame');
		ko.cleanNode($('#mainFrame')[0]);
	}

	$( '#mainFrame' ).load( './fragment/' + id + '.html?ts=' + ts, function(response, status, xhr) {
		// Deal with errors
		if ( status == "error" ) {
			if (xhr.status == 0) {
				toastr.error("Unable to reach server; try again.");
				return
			}
			var msg = "Sorry but there was an error: ";
			toastr.error( msg + xhr.status + " " + xhr.statusText );
			return;
		}

		$( '#mainFrame' ).hide( );
		console.log('Page fragment ' + id + ' load completed.');
		$( '#nav-title' ).html( $( 'H1.title' ).html() );
		$( '#mainFrame' ).show( 'slow' );

		// Nav changes -- if there are any
		window.location.hash = origId;
		selectMenu( id );
	});
} // end function loadPage

function selectMenu( item ) {
	$( '#navbar UL.navbar-nav LI' ).removeClass('active');
	$( '#navbar UL.navbar-nav LI.page-' + item.replace('.', '-') ).addClass('active');
} // end function selectMenu


////////// jQuery Extensions //////////

window.jQuery.ApiDELETE = function(apipath, successFunc) {
        $.ajax({
                url: apiBase + apipath,
                method: "DELETE",
		cache: false,
                contentType: "application/json",
                error: displayError,
                success: successFunc
        });
};

window.jQuery.ApiGET = function(apipath, successFunc) {
        $.ajax({
                url: apiBase + apipath,
                method: "GET",
		cache: false,
                contentType: "application/json",
                error: displayError,
                success: successFunc
        });
};

window.jQuery.ApiPOST = function(apipath, data, successFunc) {
        $.ajax({
                url: apiBase + apipath,
                method: "POST",
		cache: false,
		data: data,
                contentType: "application/json",
                error: displayError,
                success: successFunc
        });
};


////////// Knockout Extensions //////////

var KO_FIELD_STRING  = 0;
var KO_FIELD_NUMERIC = 1;

function koValid(field, fieldType) {
	var value = ko.utils.unwrapObservable( field );

	console.log("koValid(" + field + ", " + fieldType + ")");
	if (typeof value === 'undefined') {
		console.log("koValid(): field == undefined");
		return false;
	}
	if (fieldType == KO_FIELD_STRING) {
		console.log("koValid(): KO_FIELD_STRING");
		if (value == "") {
			console.log("koValid(): KO_FIELD_STRING -- empty");
			return false;
		}
	}
	if (fieldType == KO_FIELD_NUMERIC) {
		console.log("koValid(): KO_FIELD_NUMERIC");
		if (value < 1) {
			console.log("koValid(): KO_FIELD_NUMERIC -- < 1");
			return false;
		}
	}

	// By default
	return true;
} // end koValid

// http://jsfiddle.net/rniemeyer/sHB9p/
var koOptionsProvider = (function () {
	"use strict";
	var self = {};

	// Container for options data, a sort of dictionary of option arrays.
	self.options = {};
    
	self.init = function (optionData) {
		// Pre-populate any provided options data here...
	};

	self.reset = function( name ) {
		self.options[name] = ko.observableArray([ ]);
		return self.options[name];
	};

	self.get = function( opts ) {
		if (!self.options[opts.name]) {
			self.options[opts.name] = ko.observableArray([{ value: opts.initialValue }]);
		}

		var apiUrl = opts.apiUrl;

		// Interpolate parameter, if there is one
		if ( apiUrl.indexOf( '*' ) != -1 ) {
			apiUrl = apiUrl.replace( '*', opts.valueObservable() );
		}
            
		$.ApiGET( opts.apiUrl, function( data ) {
			self.options[opts.name]( data );
		});

		// Return reference to observable immediately.
		return self.options[opts.name];
	};
    
	return self;
})();

ko.bindingHandlers.lazyOptions = {
	init: function(element, valueAccessor, allBindings) {
		var initialized = false;
		var focused = ko.observable();

		// Create a new computed to represent our options
		var options = ko.computed({
			disposeWhenNodeIsRemoved: element,
			read: function() {
				var value;

				// Before focus, return an array with a single
				// option that matches the current value
				if (!initialized && !focused()) {
					// Determine the value by looking at
					// what the value binding is bound
					// against
					value = allBindings.get("value")();
					return [value];
				}

				// Otherwise return the actual options
				initialized = true;
				return ko.unwrap(valueAccessor());
			}
		}); // options

		// Apply the options binding with sub-options
		ko.applyBindingsToNode(element, { 
			hasFocus: focused,
			options: options,
			optionsCaption: allBindings.get("optionsCaption")
		}); 
	} // init
}; // ko.bindingHandlers.lazyOptions

// Select2 KO support from https://github.com/select2/select2/wiki/Knockout.js-Integration
ko.bindingHandlers.select2 = {
    init: function(el, valueAccessor, allBindingsAccessor, viewModel) {
      ko.utils.domNodeDisposal.addDisposeCallback(el, function() {
        $(el).select2('destroy');
      });

      var allBindings = allBindingsAccessor(),
          select2 = ko.utils.unwrapObservable(allBindings.select2);

      $(el).select2(select2);
    },
    update: function (el, valueAccessor, allBindingsAccessor, viewModel) {
        var allBindings = allBindingsAccessor();

        if ("value" in allBindings) {
            if ((allBindings.select2.multiple || el.multiple) && allBindings.value().constructor != Array) {                
                $(el).val(allBindings.value().split(',')).trigger('change');
            }
            else {
                $(el).val(allBindings.value()).trigger('change');
            }
        } else if ("selectedOptions" in allBindings) {
            var converted = [];
            var textAccessor = function(value) { return value; };
            if ("optionsText" in allBindings) {
                textAccessor = function(value) {
                    var valueAccessor = function (item) { return item; }
                    if ("optionsValue" in allBindings) {
                        valueAccessor = function (item) { return item[allBindings.optionsValue]; }
                    }
                    var items = $.grep(allBindings.options(), function (e) { return valueAccessor(e) == value});
                    if (items.length == 0 || items.length > 1) {
                        return "UNKNOWN";
                    }
                    return items[0][allBindings.optionsText];
                }
            }
            $.each(allBindings.selectedOptions(), function (key, value) {
                converted.push({id: value, text: textAccessor(value)});
            });
            $(el).select2("data", converted);
        }
        $(el).trigger("change");
    }
}; // ko.bindingHandlers.select2

function asyncComputed(evaluator, owner) {
	var result = ko.observable();
	ko.computed(function() {
	// Get the $.Deferred value, and then set up a callback so that when it's done,
	// the output is transferred onto our "result" observable
		evaluator.call(owner).done(result);
	});
	return result;
} // asyncComputed

ko.bindingHandlers.numberInput = {
	init: function (element, valueAccessor, allBindingsAccessor) {
		var value = valueAccessor();
		element.value = value();
		element.onchange = function () {            
			var strValue = this.value;
			var numValue = Number(strValue);
			numValue = isNaN(numValue) ? 0 : numValue;
			this.value = numValue;
			value(numValue);
		};
	}    
};


////////// Miscellaneous Functions //////////

function sanitizeId( orig ) {
	return String( orig )
		.replace( / /g,  '_' )
		.replace( /\//g, '_' )
		.replace( /"/g,  '_' )
		.replace( /'/g,  '_' );
} // end function sanitizeId

// Generate a password string
function randString(id){
	var dataSet = $(id).attr('data-character-set').split(',');  
	var possible = '';
	if($.inArray('a-z', dataSet) >= 0){
		possible += 'abcdefghijklmnopqrstuvwxyz';
	}
	if($.inArray('A-Z', dataSet) >= 0){
		possible += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
	}
	if($.inArray('0-9', dataSet) >= 0){
		possible += '0123456789';
	}
	if($.inArray('#', dataSet) >= 0){
		possible += '![]{}()%&*$#^<>~@|';
	}
	var text = '';
	for(var i=0; i < $(id).attr('data-size'); i++) {
		text += possible.charAt(Math.floor(Math.random() * possible.length));
	}
	return text;
}
