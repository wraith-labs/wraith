var current_panel_login_token;
var current_panel_crypt_key;
var trusted_server_signature;
var base_uri;
const window = self;

importScripts("crypto.js");

self.addEventListener("message", function(e) {
	var args = e.data;
	current_panel_login_token = args["current_panel_login_token"];
	current_panel_crypt_key = args["current_panel_crypt_key"];
	trusted_server_signature = args["trusted_server_signature"];
	base_uri = args["base_uri"];
}, false);

// Wait until we get the args to avoid race conditions
//while (true) { if (base_uri != undefined) { break; } }

function api(args) {
	console.log("Making API Call")
	args["panel_token"] = current_panel_login_token;

	const Http = new XMLHttpRequest();
	const api_url = base_uri+"api.php";
	const key = current_panel_crypt_key;
	const args_json = JSON.stringify(args);
	
	const crypt_prefix_charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890";
	var crypt_prefix = "";
	for (let i = 0; i < 5; i++) {
		crypt_prefix += crypt_prefix_charset[Math.floor(Math.random() * crypt_prefix_charset.length)];
	}
	crypt_prefix += "Wr7H"
	
	var response = "";
	var response_dict = {};
	
	Http.onreadystatechange = function() {
		if (this.readyState == 4 && this.status == 200) {
			// The response could be encrypted but it might not be
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
		// If the API returns an error in the form of a HTTP code
		} else if (this.readyState == 4 && this.status != 200) {
			console.log("The API returned a non-200 code when called. The code was: " + Http.status);
		}
	}
	
	Http.open("POST", api_url, false);
	Http.send(crypt_prefix+aes.encrypt(args_json, key));

	return response_dict;
}

// Update page every 2 seconds
(function update_page() {
	var getinfo_response;
	var getwraiths_response;
	var getsettings_response;
	try {
		// Get info for info page
		var info_page_data = api({"message_type": "getinfo"});
		postMessage(["info",info_page_data]);
		// Get info about wraiths for wraiths page
		var wraiths_page_data = api({"message_type": "getwraiths"});
		postMessage(["wraiths",info_page_data]);
		// Get a list of options for the options page
		var settings_page_data = api({"message_type": "settings", "data": "get"});
		postMessage(["settings",info_page_data]);

	} catch (err) { console.log("Error while updating page: " + err.message) }

	setTimeout(update_page, 2000);
})();
