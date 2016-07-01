// app.js

var apiBase = "../api";
var currentPage = null;
var sessionId = null;

function sanitizeId( orig ) {
	return String( orig )
		.replace( / /g,  '_' )
		.replace( /\//g, '_' )
		.replace( /"/g,  '_' )
		.replace( /'/g,  '_' );
} // end function sanitizeId

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

// $.ApiGet() wraps an API call with proper authentication and error handling.
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

$(document).ready(function() {
	console.log('document.ready');

	// Force toaster notifications to show how long until they go away
	toastr.options.progressBar = true;

	// All preperatory stuff
	$( '.nav-authed' ).hide();

	$( '#login-form' ).submit(function() {
		console.log('login-submit.click');
		login();
		return false;
	});

	if (!authenticated()) {
		console.log('!auth');
		loadPage('login-splash');
		$( '#loginDialog' ).modal('show');
		$( "#login-username" ).focus();
	}

});
