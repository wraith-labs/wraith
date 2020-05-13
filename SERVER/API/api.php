<?php
/*
The API returns JSON responses when requests are made to it. If it detects that
the client is capable of encrypted communication using the Wraith/HTTP protocol it
will automatically encrypt its replies.
*/

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

        global $db, $crypt, $cryptKey;

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

        // Close the database connection when exiting
        $db = NULL;

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

    // To keep all stats up to-date, and avoid performing actions on disconnected
    // Wraiths, expire any that have not had a heartbeat in a while first.
    dbExpireWraiths();

    // Expire any panel sessions which have not had a heartbeat recently for
    // security and to prevent sessions from sticking around because a user
    // forgot to log out.
    dbExpireSessions();

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
    foreach ($API_USERS as $id => $user) {

        // If the username exists in the database
        if ($credentials[0] === $user["userName"]) {

            // Check whether the password matches
            if (password_verify($credentials[1], $user["userPassword"])) {

                // If the username exists and matches the password,
                // create a session for the user
                $sessionID = dbCreateSession($user["userName"]);

                // Get the information of the session
                $session = dbGetSessions()[$sessionID];

                $response = [
                    "status" => "SUCCESS",
                    "config" => [
                        "sessionID" => $sessionID,
                        "sessionInfo" => $session
                    ],
                ];
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

    // Define the API version
    define("API_VERSION", "4.0.0");
    // Define an array of supported protocol versions. This will be updated
    // by each included protocol file.
    $SUPPORTED_PROTOCOL_VERSIONS = [];

    // Import some helper scripts
    require_once("helpers/db.php");      // Database access and management
    require_once("helpers/crypto.php");  // Encryption and decryption
    require_once("helpers/misc.php");    // Miscellaneous
    // Import protocol handlers
    foreach (glob("helpers/protocols/proto_v_*.php") as $protoHandler) { include($protoHandler); }

    // To keep all stats up to-date, and avoid performing actions on disconnected
    // Wraiths, expire any that have not had a heartbeat in a while first.
    dbExpireWraiths();

    // Expire any panel sessions which have not had a heartbeat recently for
    // security and to prevent sessions from sticking around because a user
    // forgot to log out.
    dbExpireSessions();

    // Define a function to respond to the client
    function respond($response) {

        global $db, $crypt, $cryptKey, $SETTINGS;

        // Set the text/plain content type header so proxies and browsers
        // don't try interpreting responses
        header("Content-Type: text/plain");
        // If a global $crypt object is defined as well as
        // a global $cryptKey, automatically encrypt the
        // response and add the prefix
        if (isset($crypt) && isset($cryptKey)) {

            $message = $SETTINGS["APIPrefix"] . $crypt->encrypt(json_encode($response), $cryptKey);

        } else {

            $message = json_encode($response);

        }

        // Close the database connection when exiting
        $db = NULL;

        // Finally, send the response and exit
        die($message);
    }

    // Check if the requesting IP is blacklisted. If so, reject the request
    $requesterIP = getClientIP();
    $IPBlacklist = json_decode($SETTINGS["requestIPBlacklist"]);
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
    if (strpos($reqBody, $SETTINGS["APIPrefix"]) !== 0) {

        $response = [
            "status" => "ERROR",
            "message" => "incorrectly formatted request",
        ];
        respond($response);

    }

    // The character following this should be a non-zero, integer.
    // It indicates whether the client is a Wraith or panel:
    // Odd == Wraith
    // Even == Panel
    // First, check if the character after the prefix is a non-zero integer
    $reqIdentificationChar = $reqBody[strlen($SETTINGS["APIPrefix"])];
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

        // Even - the request is coming from a panel
        $requester = "panel";

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
    $reqProtocolvChar = $reqBody[strlen($SETTINGS["APIPrefix"])+1];
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
    // Wraith or panel, as well as the protocol version in use, we can get rid of
    // the header from the message.
    $reqBody = substr($reqBody, strlen($SETTINGS["APIPrefix"])+2);

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
        $data = $crypt->decrypt($reqBody, $SETTINGS["wraithSwitchCryptKey"]);

        // Try JSON decoding the decrypted data
        $data = json_decode($data, true);
        // If this failed, either the data was not JSON, was encrypted with
        // a different key, or is invalid
        if ($data === null) {

            // In case we have used the wrong key, try the default key instead
            $data = $crypt->decrypt($reqBody, $SETTINGS["wraithInitialCryptKey"]);

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
                $cryptKey = $SETTINGS["wraithInitialCryptKey"];

            }

        } else {

            // If this worked, use the switch key for the response from now on
            $cryptKey = $SETTINGS["wraithSwitchCryptKey"];

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

    } elseif ($requester === "panel") {

        // TODO

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
        "SETTINGS",
        "db",
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
    $handler = new $handlerClassName($db, $requester, $requesterIP, $data, $SETTINGS);

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
