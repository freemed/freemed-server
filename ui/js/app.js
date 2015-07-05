// app.js

var apiBase = "../api";
var currentPage = null;
var sessionId = null;

function authenticated() {
	return sessionId != null;
} // end function authenticated

function loadPage( id ) {
	console.log('Loading page ' + id);
	$( '#container' ).load( './' + id + '.html', function() {
		console.log('Page fragment load completed.');
	});
} // end function loadPage

function login() {
	$.ajax({
		url: apiBase + "/auth/login",
		method: "POST",
		contentType: "application/json",
		data: JSON.stringify({
			user: $('#login-username').val(),
			pass: $('#login-password').val()
		}),
		error: function(x){
			console.log(JSON.stringify(x));
			loginStateChange(false, null);
		},
		success: function(data){
			sessionId = data.session_id;
			loginStateChange(true, function() {
				console.log('cb: sessionId = ' + sessionId);
				if (currentPage == null || currentPage != 'login-splash') {
					loadPage('main');
				}
			});
		}
	});
} // end function login

function loginStateChange(loggedin, cb) {
	if (loggedin) {
		console.log('Login successful');
		$( '#loginDialog' ).modal('hide');
		if (cb != null) {
			cb();
		}
	} else {
		console.log('Login failed');
		$( '#loginDialog' ).modal('show');
	}
} // end function loginStateChange

$(document).ready(function() {
	console.log('document.ready');

	$('#login-form').submit(function() {
		console.log('login-submit.click');
		login();
		return false;
	});

	if (!authenticated()) {
		console.log('!auth');
		//loadPage('login-splash');
		$( '#loginDialog' ).modal('show');
	}
});
