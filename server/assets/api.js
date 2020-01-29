// A function which sends requests to the API and returns a processed response
function api(args) {
	// Add the panel token to the response to authenticate the response
	args["panel_token"] = current_panel_login_token;

	// Define some constants
	const Http = new XMLHttpRequest();
	const api_url = base_uri+"api.php";
	const key = current_panel_crypt_key;
	const args_json = JSON.stringify(args);
	
	// Generate a prefix by which the API can identify that the request is from a valid source
	const crypt_prefix_charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890";
	var crypt_prefix = "";
	for (let i = 0; i < 5; i++) {
		crypt_prefix += crypt_prefix_charset[Math.floor(Math.random() * crypt_prefix_charset.length)];
	}
	crypt_prefix += "Wr7H"
	
	// Define the variables to hold the response from the API
	var response = "";
	var response_dict = {};
	
	// When a response is received
	Http.onreadystatechange = function() {
		// If the response was received successfully
		if (this.readyState == 4 && this.status == 200) {
			// The response could be encrypted but it might not be
			try {
				// Try decrypting the response
		 		response = aes.decrypt(Http.responseText, key);
			 	// Parse the JSON response
			 	response_dict = JSON.parse(response);
		 		// Verify the fingerprint of the server (only if the communication is encrypted because otherwise the server does not attach it)
		 		if (response_dict["server_id"] != trusted_server_signature) {
		 			const untrusted_server_error = "The server provided an incorrect ID. This should never happen unless something went very wrong. The server ID is `" + response_dict["server_id"] + "` while the expected ID is `" + trusted_server_signature + "`. Quitting."
		 			throw new Error(untrusted_server_error);
		 		}
		 	} catch (err) {
		 		// If we got here, the response was not encrypted or the server provided the wrong ID
		 		// Id the latter, re-throw the error so the user can see
		 		if (err.message.startsWith("The server provided an incorrect ID. ")) { throw err; }
		 		// If we got here, that means it was just a simple decryption problem so get the response as plaintext
		 		response = Http.responseText;
		 		// Parse the JSON response
		 		try {response_dict = JSON.parse(response);}
		 		// If the API response is not valid JSON, log the error
		 		catch {
		 			console.log("The API response contains invalid JSON");
		 			// Set both response variables to undefined so this error can be detected down the line
		 			response = undefined;
					response_dict = undefined;
		 		}
		 	}
		 	// If the API returns a message, tell the user
		 	if ("message" in response_dict) {
		 		console.log("The server says: "+response_dict["message"]);
		 	}
		// If the API returns an error in the form of a HTTP code
		} else if (this.readyState == 4 && this.status != 200) {
			console.log("The API returned a non-200 code when called. The code was: " + Http.status);
 			// Set both response variables to undefined so this error can be detected down the line
			response = undefined;
			response_dict = undefined;
		}
	}
	
	// Actually send the request
	Http.open("POST", api_url, false);
	Http.send(crypt_prefix+aes.encrypt(args_json, key));

	// Return the response from the API to the caller
	return response_dict;
}
