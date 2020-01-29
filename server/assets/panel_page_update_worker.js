// Define window as self (very similar objects) to make sure everything works
const window = self;

// Import the encryption library for encrypted communication
importScripts("crypto.js");
// Import the api library to communicate with the API
importScripts("api.js");

// Wait for the nescessary variables
self.addEventListener("message", function(e) {
	// Get the veriables from the message
	var args = e.data;
	self.current_panel_login_token = args["current_panel_login_token"];
	self.current_panel_crypt_key = args["current_panel_crypt_key"];
	self.trusted_server_signature = args["trusted_server_signature"];
	self.base_uri = args["base_uri"];
	
	// Start the page updating loop
	update_page();
}, false);

// Update page every 5 seconds
function update_page() {
	// This function should never exit with an error because that would stop the page from updating
	// so just catch and print the errors to console
	try {

		// Get info for info page
		var info_page_data = api({"message_type": "getinfo"});
		// Only forward the data if it's actually defined
		if (info_page_data != undefined) {postMessage(["info",info_page_data])};
		// Get info about wraiths for wraiths page
		//var wraiths_page_data = api({"message_type": "getwraiths"});
		//if (wraiths_page_data != undefined) {postMessage(["wraiths",info_page_data]);}
		// Get a list of options for the options page
		//var settings_page_data = api({"message_type": "settings", "data": "get"});
		//if (settings_page_data != undefined) {postMessage(["settings",info_page_data]);}

	} catch (err) { console.log("Error in worker while updating page: " + err.message) }

	// Repeat in 5 seconds (5 seems to be a good balance between quick refreshing and conserving resources
	setTimeout(update_page, 5000);
};
