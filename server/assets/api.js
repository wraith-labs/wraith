function api(args) {
	args["panel_token"] = current_panel_login_token;

	const Http = new XMLHttpRequest();
	const api_url="/api.php";
	const key = current_panel_crypt_key;
	const args_json = JSON.stringify(args);
	
	const crypt_prefix_charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890";
	var crypt_prefix = "";
	for (let i = 0; i < 5; i++) {
		crypt_prefix += crypt_prefix_charset[Math.floor(Math.random() * crypt_prefix_charset.length)];
	}
	crypt_prefix += "Wr7H"
	
	Http.open("POST", api_url, true);
	Http.send(crypt_prefix+aes.encrypt(args_json, key));

	Http.onload = function() {
		// Make sure the API had no errors (check status code)
		if (Http.status != 200) {
			alert("The API returned a non-200 code when called. The exit code was: " + Http.status);
			return;
		}
		// The response could be encrypted but it might not be
		var response;
		var response_dict;
		try {
	 		response = aes.decrypt(Http.responseText, key);
		 	// Parse the JSON response
		 	response_dict = JSON.parse(response);
	 		// Verify the fingerprint of the server (only if the communication is encrypted because otherwise the server does not attach it)
	 		if (response_dict["server_id"] != trusted_server_signature) {
	 			const untrusted_server_error = "The server provided an incorrect ID. This should never happen unless something went very wrong. The server ID is `" + response_dict["server_id"] + "` while the expected ID is `" + trusted_server_signature + "`. Quitting."
	 			throw new Error(untrusted_server_error);
	 		}
	 	} catch (err) {
	 		if (err.message.startsWith("The server provided an incorrect ID. ")) { throw err; }
	 		response = Http.responseText;
	 		// Parse the JSON response
	 		response_dict = JSON.parse(response);
	 	}
	 	// If the API returns a message, tell the user
	 	if ("message" in response_dict) {
	 		alert("The server says: "+response_dict["message"]);
	 	}
	 	return response_dict;
	}
}
