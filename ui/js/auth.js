// auth.js: Authentication functions

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

