// auth.js: Authentication functions

var sessionAuth = function(xhr) {
	xhr.setRequestHeader('Authorization', 'Bearer ' + sessionId);
};

function authenticated() {
	return sessionId != null;
} // end function authenticated

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
		$( 'LI.nav-authed' ).show();
		$( '#loginDialog' ).modal('hide');
		if (cb != null) {
			cb();
		}
	} else {
		console.log('Login failed');
		$( 'LI.nav-authed' ).hide();
		$( '#loginDialog' ).modal('show');
	}
} // end function loginStateChange

function logout() {
	$.ajax({
		url: apiBase + "/auth/logout",
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

