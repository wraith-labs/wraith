<?php
/*
The API returns JSON responses when requests are made to it. If it detects that
the client is capable of encrypted communication using the Wraith/HTTP protocol it
will automatically encrypt its replies.
*/

// Convert all errors into catch-able exceptions
set_error_handler(function($errno, $errstr, $errfile, $errline ){

    throw new ErrorException($errstr, $errno, 0, $errfile, $errline);

});

// Define the API version
define("API_VERSION", "4.0.0");

// Add the nescessary headers to support cross-origin requests as
// API managers can be hosted anywhere.
header("Access-Control-Allow-Origin: *");
header("Access-Control-Allow-Methods: GET, PUT, POST");

// The API only uses GET, POST and PUT requests. Other request methods can be
// discarded and return an error message.

// GET requests always return the current URL. This is to allow hard-coding the
// API URL in the Wraith.
if ($_SERVER["REQUEST_METHOD"] === "GET") {

    // Function to return the full URL of the current document
    function getDocumentURL() {
        $s = &$_SERVER;
        $ssl = (!empty($s['HTTPS']) && $s['HTTPS'] == 'on') ? true:false;
        $sp = strtolower($s['SERVER_PROTOCOL']);
        $protocol = substr($sp, 0, strpos($sp, '/')) . (($ssl) ? 's' : '');
        $port = $s['SERVER_PORT'];
        $port = ((!$ssl && $port=='80') || ($ssl && $port=='443')) ? '' : ':'.$port;
        $host = isset($s['HTTP_X_FORWARDED_HOST']) ? $s['HTTP_X_FORWARDED_HOST'] : (isset($s['HTTP_HOST']) ? $s['HTTP_HOST'] : null);
        $host = isset($host) ? $host : $s['SERVER_NAME'] . $port;
        $uri = $protocol . '://' . $host . $s['REQUEST_URI'];
        $segments = explode('?', $uri, 2);
        $url = $segments[0];
        return $url;
    }

    header("Content-Type: text/plain");
    die(getDocumentURL());

// PUT requests are used by API managers (such as the panel) for automatic
// configuration. They must contain valid manager account authentication
// with every request. When a valid request is received, a session is
// created for the user account and the manager should use POST from then on.
} else if ($_SERVER["REQUEST_METHOD"] == "PUT") {

    // Manager autoconf requests do not follow the usual protocol
    // as the manager does not know the proper prefix or encryption
    // key to use. These requests should be processed differently
    // from normal POST requests.

    // The request body must contain a user's username and password.
    // The password will be obfuscated by encrypting it using the standard
    // method and using the username + "wraithCredentials" as the key.
    // THIS IS NOT SECURE! USE SSL (HTTPS) TOO!

    // Define a function to respond to the client
    function respond($response) {

        global $crypt, $cryptKey;

        // Set the text/plain content type header so proxies and browsers
        // don't try interpreting responses
        header("Content-Type: text/plain");
        // If a global $crypt object is defined as well as
        // a global $cryptKey, automatically encrypt the
        // response
        if (isset($crypt) && isset($cryptKey)) {

            $message = $crypt->encrypt(json_encode($response), $cryptKey);

        } else {

            $message = json_encode($response);

        }

        // Finally, send the response and exit
        die($message);
    }

    // As autoconf is equivalent to authentication, start with a
    // "random" time delay to mitigate brute-force attacks and
    // make it difficult to guess whether authentication was
    // successful based on the time taken for the response to
    // arrive. Brute-force attacks can still be fairly effective
    // if the attacker makes concurrent requests however.
    // Between 0.5 and 2 seconds
    usleep(rand(500000, 2000000));

    // Import the required helper scripts
    require_once("helpers/db.php");      // Database access and management
    require_once("helpers/crypto.php");  // Encryption and decryption
    require_once("helpers/misc.php");    // Miscellaneous

    // Create an instance of the database manager
    $dbm = new DBManager();

    // To keep all stats up to-date, and avoid performing actions on disconnected
    // Wraiths, expire any that have not had a heartbeat in a while first.
    $dbm->dbExpireWraiths();

    // Expire any manager sessions which have not had a heartbeat recently for
    // security and to prevent sessions from sticking around because a user
    // forgot to log out.
    $dbm->dbExpireSessions();

    // Re-generate the first-layer encryption key for management
    // sessions for better security (only if there are no active sessions)
    $dbm->dbRegenMgmtCryptKeyIfNoSessions();

    // Get the request body to verify the credentials
    $reqBody = file_get_contents("php://input");

    // The request body should have 1 instance of the `|` sign
    // separating the username and password
    if (substr_count($reqBody, "|") !== 1) {

        $response = [
            "status" => "ERROR",
            "message" => "incorrectly formatted request",
        ];
        respond($response);
    }

    // Create a crypt object to de-obfuscate the password (and later encrypt
    // the response - no responses are encrypted until $cryptKey is defined).
    $crypt = new aes();

    // Split the request into the username and password
    $credentials = explode("|", $reqBody, 2);
    // Unobfuscate the password
    $credentials[1] = $crypt->decrypt($credentials[1], $credentials[0] . "wraithCredentials");

    // Check whether the username is indeed a valid user account
    foreach ($dbm->dbGetUsers() as $id => $user) {

        // If the username exists in the database
        if ($credentials[0] === $user["userName"]) {

            // Check whether the password matches
            if (password_verify($credentials[1], $user["userPassword"])) {

                // If the username exists and matches the password,
                // create a session for the user
                $sessionID = $dbm->dbAddSession($user["userName"]);

                // Get the information of the session
                $thisSession = $dbm->dbGetSessions([
                    "assignedID" => [$sessionID]
                ])[$sessionID];

                // Encrypt the response with the password of the user. This
                // is again not too secure but we're mainly relying on SSL
                $cryptKey = $credentials[1];

                // Generate a response to be sent
                $response = [
                    "status" => "SUCCESS",
                    "config" => [
                        "sessionID" => $sessionID,
                        "sessionToken" => $thisSession["sessionToken"],
                        "updateInterval" => $dbm->dbGetSetting(["key" => ["managementSessionExpiryDelay"]])["managementSessionExpiryDelay"] / 3,
                        "APIPrefix" => $dbm->dbGetSetting(["key" => ["APIPrefix"]])["APIPrefix"],
                        "firstLayerEncryptionKey" => $dbm->dbGetSetting(["key" => ["managementFirstLayerEncryptionKey"]])["managementFirstLayerEncryptionKey"],
                        "APIVersion" => API_VERSION,
                        "APIFingerprint" => $dbm->dbGetSetting(["key" => ["APIFingerprint"]])["APIFingerprint"]
                    ],
                ];
                // ...and send it
                respond($response);

            }

        }

    }

    // If we got here, the credentials must have been incorrect
    $response = [
        "status" => "ERROR",
        "message" => "incorrect credentials",
    ];
    respond($response);

// POST requests are used for actual interaction with the API using the Wraith
// protocol. All non-compliant requests result in errors.
} else if ($_SERVER["REQUEST_METHOD"] === "POST") {

    // Define an array of supported protocol versions. This will be updated
    // by each included protocol file.
    $SUPPORTED_PROTOCOL_VERSIONS = [];

    // Import some helper scripts
    require_once("helpers/db.php");      // Database access and management
    require_once("helpers/crypto.php");  // Encryption and decryption
    require_once("helpers/misc.php");    // Miscellaneous
    // Import protocol handlers
    foreach (glob("helpers/protocols/proto_v_*.php") as $protoHandler) { include($protoHandler); }

    // Create an instance of the database manager
    $dbm = new DBManager();

    // To keep all stats up to-date, and avoid performing actions on disconnected
    // Wraiths, expire any that have not had a heartbeat in a while first.
    $dbm->dbExpireWraiths();

    // Expire any manager sessions which have not had a heartbeat recently for
    // security and to prevent sessions from sticking around because a user
    // forgot to log out.
    $dbm->dbExpireSessions();

    // Define a function to respond to the client
    function respond($response) {

        global $crypt, $cryptKey;

        // Set the text/plain content type header so proxies and browsers
        // don't try interpreting responses
        header("Content-Type: text/plain");
        // If a global $crypt object is defined as well as
        // a global $cryptKey, automatically encrypt the
        // response and add the prefix
        if (isset($crypt) && isset($cryptKey)) {

            $message = $dbm->dbGetSetting(["key" => ["APIPrefix"]])["APIPrefix"] . $crypt->encrypt(json_encode($response), $cryptKey);

        } else {

            $message = json_encode($response);

        }

        // Finally, send the response and exit
        die($message);
    }

    // Check if the requesting IP is blacklisted. If so, reject the request
    $requesterIP = getClientIP();
    $IPBlacklist = json_decode($dbm->dbGetSetting(["key" => ["requestIPBlacklist"]])["requestIPBlacklist"]);
    if (in_array($requesterIP, $IPBlacklist)) {

        $response = [
            "status" => "ERROR",
            "message" => "you have been blocked from accessing this resource",
        ];
        respond($response);

    }

    // Get the request body
    $reqBody = file_get_contents("php://input");

    /*

    REQUEST VALIDATION AND PREPARATION

    */

    // Find if the request starts with the pre-defined prefix. If not,
    // it is invalid.
    if (strpos($reqBody, $dbm->dbGetSetting(["key" => ["APIPrefix"]])["APIPrefix"]) !== 0) {

        $response = [
            "status" => "ERROR",
            "message" => "incorrectly formatted request",
        ];
        respond($response);

    }

    // The character following this should be a non-zero, integer.
    // It indicates whether the client is a Wraith or manager:
    // Odd == Wraith
    // Even == Manager
    // First, check if the character after the prefix is a non-zero integer
    $reqIdentificationChar = $reqBody[strlen($dbm->dbGetSetting(["key" => ["APIPrefix"]])["APIPrefix"])];
    if ($reqIdentificationChar % 10 === 0) {

        // 0 is only returned by non-integer characters or 0 itself
        // both of which aren't valid (or multiples of 10
        // but we're only getting a single digit/char)
        $response = [
            "status" => "ERROR",
            "message" => "incorrectly formatted request",
        ];
        respond($response);

    } elseif ($reqIdentificationChar % 2 === 1) {

        // Odd - the request is coming from a Wraith
        $requester = "wraith";

    } elseif ($reqIdentificationChar % 2 === 0) {

        // Even - the request is coming from a manager
        $requester = "manager";

    } else {

        // This should never happen but better safe than sorry
        $response = [
            "status" => "ERROR",
            "message" => "incorrectly formatted request",
        ];
        respond($response);

    }

    // The next char indicates the protocol version that the requester is using.
    // This should be checked against the protocol versions we support and the
    // request should be rejected if the protocol is unsupported.
    $reqProtocolvChar = $reqBody[strlen($dbm->dbGetSetting(["key" => ["APIPrefix"]])["APIPrefix"])+1];
    if (!(in_array($reqProtocolvChar, $SUPPORTED_PROTOCOL_VERSIONS))) {

        $response = [
            "status" => "ERROR",
            "message" => "unsupported protocol",
        ];
        respond($response);

    } else {

        $protocolVersion = $reqProtocolvChar;

    }

    // Now that we know that the request is valid and whether it comes from a
    // Wraith or manager, as well as the protocol version in use, we can get rid of
    // the header from the message.
    $reqBody = substr($reqBody, strlen($dbm->dbGetSetting(["key" => ["APIPrefix"]])["APIPrefix"])+2);

    // The remainder of the request should simply be the encrypted data
    // This needs to be decrypted using the correct decryption key which
    // we find below

    // First, define the encryption object, but not the $cryptKey variable.
    // Without the $cryptKey variable set, respond() will not encrypt responses
    // in case there is an error with decryption
    $crypt = new aes();

    if ($requester === "wraith") {

        // The decryption key will either be the default (if the Wraith is not
        // logged in) or the switch key. The only way to tell those apart is to
        // try decrypting with both.

        // Decrypt with the switch key first as this will be used more often
        $data = $crypt->decrypt($reqBody, $dbm->dbGetSetting(["key" => ["wraithSwitchCryptKey"]])["wraithSwitchCryptKey"]);

        // Try JSON decoding the decrypted data
        $data = json_decode($data, true);
        // If this failed, either the data was not JSON, was encrypted with
        // a different key, or is invalid
        if ($data === null) {

            // In case we have used the wrong key, try the default key instead
            $data = $crypt->decrypt($reqBody, $dbm->dbGetSetting(["key" => ["wraithInitialCryptKey"]])["wraithInitialCryptKey"]);

            // Try JSON decoding the decrypted data
            $data = json_decode($data, true);
            // If this failed, either the data must be invalid JSON or invalid altogether
            if ($data === null) {

                // In both cases, we return an error message
                $response = [
                    "status" => "ERROR",
                    "message" => "incorrectly formatted request",
                ];
                respond($response);

            } else {

                // If this worked, use the default key for the response from now on
                $cryptKey = $dbm->dbGetSetting(["key" => ["wraithInitialCryptKey"]])["wraithInitialCryptKey"];

            }

        } else {

            // If this worked, use the switch key for the response from now on
            $cryptKey = $dbm->dbGetSetting(["key" => ["wraithSwitchCryptKey"]])["wraithSwitchCryptKey"];

        }

        // At this point, we should have a variable named $data holding
        // data received from the Wraith.

        // We just need to make sure that the data is an array and the required
        // headers are present

        // Check if the data is an array
        if (!(is_array($data))) {

            $response = [
                "status" => "ERROR",
                "message" => "incorrectly formatted request",
            ];
            respond($response);

        }

        // Make sure the array has the required keys
        if (!(hasKeys($data, [
            "reqType", // So we know what to do with the request
        ]))) {

            $response = [
                "status" => "ERROR",
                "message" => "incorrectly formatted request",
            ];
            respond($response);

        }

        // The Wraith's request has now been fully validated and prepared and can
        // be processed

    } elseif ($requester === "manager") {

        // The decryption key for the whole request body in the case of a
        // manager will be the management first layer encryption key in
        // the settings table in the database. This key is generated when
        // the database is first created and re-generated on each valid
        // autoconf request when no sessions are active to increase security.
        // The key is sent to the manager on successful login (autoconf).
        // The decrypted body of the request will have a session ID
        // as well as another encrypted payload. This payload can be decrypted
        // using the session token attached to the session ID in the database.

        // Decrypt with the first layer key
        $data = $crypt->decrypt($reqBody, $dbm->dbGetSetting(["key" => ["managementFirstLayerEncryptionKey"]])["managementFirstLayerEncryptionKey"]);

        // Try JSON decoding the decrypted data
        $data = json_decode($data, true);
        // If this failed, the data is invalid (encrypted with an invalid key
        // or not encrypted)
        if ($data === null) {

            $response = [
                "status" => "ERROR",
                "message" => "incorrectly formatted request",
            ];
            respond($response);

        }

        // Check that the result is an array
        if (!(is_array($data))) {

            $response = [
                "status" => "ERROR",
                "message" => "incorrectly formatted request",
            ];
            respond($response);

        }

        // Make sure the array has exactly two elements
        if (sizeof($data) !== 2) {

            $response = [
                "status" => "ERROR",
                "message" => "incorrectly formatted request",
            ];
            respond($response);

        }

        // Make sure that the first element of the array is a valid session ID
        $requestSessionID = $data[0];
        $thisSession = $dbm->dbGetSessions(["assignedID" => [$requestSessionID]]);
        if (!(array_key_exists($requestSessionID, $thisSession))) {

            $response = [
                "status" => "ERROR",
                "message" => "invalid session data",
            ];
            respond($response);

        }

        // For ease of use, move the session data one level up
        $thisSession = $thisSession[$requestSessionID];

        // Get the session token associated with the session ID
        $sessionToken = $thisSession["sessionToken"];

        // Decrypt the second layer
        $data = $crypt->decrypt($data[1], $sessionToken);

        // Try JSON decoding the decrypted data
        $data = json_decode($data, true);

        // If this failed, the data is invalid
        if ($data === null) {

            $response = [
                "status" => "ERROR",
                "message" => "invalid session data",
            ];
            respond($response);

        }

        // Set the encryption key to the session token
        $cryptKey = $sessionToken;

        // Ensure that the new data is an array
        if (!(is_array($data))) {

            $response = [
                "status" => "ERROR",
                "message" => "incorrectly formatted request",
            ];
            respond($response);

        }

        // Make sure the array has the required keys
        if (!(hasKeys($data, [
            "reqType", // So we know what to do with the request
            "sessionToken", // To catch decryption errors and key collisions
        ]))) {

            $response = [
                "status" => "ERROR",
                "message" => "incorrectly formatted request",
            ];
            respond($response);

        }

        // Make sure that the session token matches
        if ($data["sessionToken"] !== $sessionToken) {

            $response = [
                "status" => "ERROR",
                "message" => "invalid session data",
            ];
            respond($response);

        }

        // Add the session ID to the data array for easier processing
        // by the protocol handler
        $data["sessionID"] = $requestSessionID;

        // The manager's request has now been fully validated and prepared and can
        // be processed

    } else {

        // This will never happen if the code is unmodified. However, to gracefully
        // handle mistakes in modification, this should stay here
        $response = [
            "status" => "ERROR",
            "message" => "the request was identified but methods for handling it were not implemented",
        ];
        respond($response);

    }

    // Unset everything other than the required variables to save resources and namespace
    $keepVariables = [
        // Superglobals
        "_ENV",
        "_GET",
        "_POST",
        "_SERVER",
        "_COOKIE",
        "_FILES",
        "_SESSION",
        // Other needed variables
        "GLOBALS",
        "dbm",
        "data",
        "requester",
        "requesterIP",
        "protocolVersion",
        "crypt",
        "cryptKey",
        // This variable
        "keepVariables"
    ];

    foreach (get_defined_vars() as $name => $value) {

        if (!(in_array($name, $keepVariables))) {

            unset($$name);

        }

    }
    unset($keepVariables, $name, $value);

    /*

    REQUEST PROCESSING

    */

    // Create an instance of the handler class for the specified protocol
    $handlerClassName = "Handler_proto_v_".$protocolVersion;
    $handler = new $handlerClassName($dbm, $requester, $requesterIP, $data);

    // Handle the request using the created handler
    $handler->handleRequest();

    // Unset (destroy) the handler so it can clean up and respond to the
    // requester
    unset($handler);

    // If nothing was sent until now, assume that there was some
    // kind of error, most likely due to a faulty protocol handler
    $response = [
        "status" => "ERROR",
        "message" => "no response generated",
    ];
    respond($response);

// OPTIONS requests are sent by browsers to check if they should allow
// cross-origin requests. As the API manager can be under any URL, all
// origins should be accepted. This however is done at the top of the
// script anyway so nothing needs to be done here.
} else if ($_SERVER["REQUEST_METHOD"] === "OPTIONS") {

    die();

} else {

    http_response_code(405);
    header("Content-Type: text/plain");
    header("Allow: GET, POST");
    die("Unsupported method");

}
