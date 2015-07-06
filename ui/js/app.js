// app.js

var apiBase = "../api";
var currentPage = null;
var sessionId = null;

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
	$( '#container' ).hide( 'slow' );
	$( '#container' ).load( './' + id + '.html?ts=' + ts, function() {
		console.log('Page fragment load completed.');
		$( '#nav-title' ).html( $( 'H1.title' ).html() );
		$( '#container' ).show( 'slow' );

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

$(document).ready(function() {
	console.log('document.ready');

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
