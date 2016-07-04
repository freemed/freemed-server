// app.js

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
