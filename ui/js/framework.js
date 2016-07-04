// framework.js: jQuery + Knockout + Bootstrap framework
// @jbuchbinder


////////// Global Variable and Settings //////////


var apiBase = "../api";
var currentPage = null;
var sessionId = null;


////////// Authentication Functions //////////


var sessionAuth = function(xhr) {
	xhr.setRequestHeader('Authorization', 'Bearer ' + sessionId);
};

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

	var ts = new Date().getTime();

	if (id != 'login-splash') {
		console.log('Removing bindings from #mainFrame');
		ko.cleanNode($('#mainFrame')[0]);
	}
	$( '#mainFrame' ).hide( );

	$( '#mainFrame' ).load( './' + id + '.html?ts=' + ts, function() {
		console.log('Page fragment load completed.');
		$( '#nav-title' ).html( $( 'H1.title' ).html() );
		$( '#mainFrame' ).show( 'slow' );

		// Deal with errors
		if ( status == "error" ) {
			var msg = "Sorry but there was an error: ";
			displayError( msg + xhr.status + " " + xhr.statusText );
			return;
		}

		// Nav changes -- if there are any
		selectMenu( id );
	});
} // end function loadPage

function selectMenu( item ) {
	$( '#navbar UL.navbar-nav LI' ).removeClass('active');
	$( '#navbar UL.navbar-nav LI.page-' + item ).addClass('active');
} // end function selectMenu


////////// jQuery Extensions //////////


window.jQuery.ApiDELETE = function(apipath, successFunc) {
        $.ajax({
                url: apiBase + apipath,
                method: "DELETE",
                contentType: "application/json",
                beforeSend: sessionAuth,
                error: displayError,
                success: successFunc
        });
};

window.jQuery.ApiGET = function(apipath, successFunc) {
        $.ajax({
                url: apiBase + apipath,
                method: "GET",
                contentType: "application/json",
                beforeSend: sessionAuth,
                error: displayError,
                success: successFunc
        });
};

window.jQuery.ApiPOST = function(apipath, data, successFunc) {
        $.ajax({
                url: apiBase + apipath,
                method: "POST",
		data: data,
                contentType: "application/json",
                beforeSend: sessionAuth,
                error: displayError,
                success: successFunc
        });
};


////////// Miscellaneous Functions //////////


function sanitizeId( orig ) {
	return String( orig )
		.replace( / /g,  '_' )
		.replace( /\//g, '_' )
		.replace( /"/g,  '_' )
		.replace( /'/g,  '_' );
} // end function sanitizeId

