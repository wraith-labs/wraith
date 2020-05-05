<?php
/*
The API returns JSON responses when requests are made to it. If it detects that
the client is capable of encrypted communication using the Wraith/HTTP protocol it
will automatically encrypt its replies.
*/

// The API only uses POST requests so assume GET requests are from Wraiths with
// a hard-coded API URL (this assumption does not disclose any sensitive
// information as it just returns the URL of the API which of course whoever is
// connecting already has) and discard any other methods.
if ($_SERVER['REQUEST_METHOD'] == "GET") {

    // Function to return the full URL of the current document
    function get_document_url() {
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
    die(get_document_url());

} else if ($_SERVER['REQUEST_METHOD'] == "POST") {

    // Define the API version
    define("API_VERSION", "4.0.0");
    // Define an array of supported protocol versions. This will be updated
    // by each included protocol file.
    $SUPPORTED_PROTOCOL_VERSIONS = [];

    // Import some helper scripts
    require("helpers/db.php");      // Database access and management
    require("helpers/crypto.php");  // Encryption and decryption
    require("helpers/misc.php");    // Miscellaneous
    // Import protocol handlers
    foreach (glob("helpers/protocols/proto_v_*.php") as $proto_handler) { include($proto_handler); }

    // To keep all stats up to-date, and avoid performing actions on disconnected
    // Wraiths, expire any that have not had a heartbeat in a while first.
    db_expire_wraiths();

    // Define a function to respond to the client
    function respond($response) {

        global $db, $crypt, $crypt_key, $SETTINGS;

        // Set the text/plain content type header so proxies and browsers
        // don't try interpreting responses
        header("Content-Type: text/plain");
        // If a global $crypt object is defined as well as
        // a global $crypt_key, and encryption is not disabled,
        // automatically encrypt the response and add the prefix
        if (isset($crypt) && isset($crypt_key) && $SETTINGS["NoEncrypt"] === "0") {

            global $crypt;
            global $crypt_key;
            $message = $SETTINGS["APIPrefix"] . $crypt->encrypt(json_encode($response), $crypt_key);

        } else {

            $message = json_encode($response);

        }

        // Close the database connection when exiting
        $db = NULL;

        // Finally, send the response and exit
        die($message);
    }

    // Check if the requesting IP is blacklisted. If so, reject the request
    $requester_IP = get_client_ip();
    $IP_blacklist = json_decode($SETTINGS["RequestIPBlacklist"]);
    if (in_array($requester_IP, $IP_blacklist)) {

        $response = [];
        $response["status"] = "ERROR";
        $response["message"] = "you have been blocked from accessing this resource";
        respond($response);

    }

    // Get the request body
    $req_body = file_get_contents("php://input");

    /*

    REQUEST VALIDATION AND PREPARATION

    */

    // Find if the request starts with the pre-defined prefix. If not,
    // it is invalid.
    if (strpos($req_body, $SETTINGS["APIPrefix"] !== 0)) {

        $response = [];
        $response["status"] = "ERROR";
        $response["message"] = "incorrectly formatted request";
        respond($response);

    }

    // The character following this should be a non-zero, integer.
    // It indicates whether the client is a Wraith or panel:
    // Odd == Wraith
    // Even == Panel
    // First, check if the character after the prefix is a non-zero integer
    $req_identification_char = $req_body[strlen($SETTINGS["APIPrefix"])];
    if ($req_identification_char % 10 === 0) {

        // 0 is only returned by non-integer characters or 0 itself
        // both of which aren't valid (or multiples of 10
        // but we're only getting a single digit/char)
        $response = [];
        $response["status"] = "ERROR";
        $response["message"] = "incorrectly formatted request";
        respond($response);

    } elseif ($req_identification_char % 2 === 1) {

        // Odd - the request is coming from a Wraith
        $requester = "wraith";

    } elseif ($req_identification_char % 2 === 0) {

        // Even - the request is coming from a panel
        $requester = "panel";

    } else {

        // This should never happen but better safe than sorry
        $response = [];
        $response["status"] = "ERROR";
        $response["message"] = "incorrectly formatted request";
        respond($response);

    }

    // The next char indicates the protocol version that the requester is using.
    // This should be checked against the protocol versions we support and the
    // request should be rejected if the protocol is unsupported.
    $req_protocolv_char = $req_body[strlen($SETTINGS["APIPrefix"])+1];
    if (!(in_array($req_protocolv_char, $SUPPORTED_PROTOCOL_VERSIONS))) {

        $response = [];
        $response["status"] = "ERROR";
        $response["message"] = "unsupported protocol";
        respond($response);

    } else {

        $protocol_version = $req_protocolv_char;

    }

    // Now that we know that the request is valid and whether it comes from a
    // Wraith or panel, as well as the protocol version in use, we can get rid of
    // the header from the message.
    $req_body = substr($req_body, strlen($SETTINGS["APIPrefix"])+2);

    // The remainder of the request should simply be the encrypted data
    // This needs to be decrypted using the correct decryption key which
    // we find below

    // First, define the encryption object, but not the $crypt_key variable.
    // Without the $crypt_key variable set, respond() will not encrypt responses
    // in case there is an error with decryption
    $crypt = new aes();

    if ($requester === "wraith") {

        // The decryption key will either be the default (if the Wraith is not
        // logged in) or the switch key. The only way to tell those apart is to
        // try decrypting with both.

        // Decrypt with the switch key first as this will be used more often
        $data = $crypt->decrypt($req_body, $SETTINGS["WraithSwitchCryptKey"]);

        // Try JSON decoding the decrypted data
        $data = json_decode($data, true);
        // If this failed, either the data was not JSON, was encrypted with
        // a different key, or is invalid
        if ($data === null) {

            // In case we have used the wrong key, try the default key instead
            $data = $crypt->decrypt($req_body, $SETTINGS["WraithInitialCryptKey"]);

            // Try JSON decoding the decrypted data
            $data = json_decode($data, true);
            // If this failed, either the data must be invalid JSON or invalid altogether
            if ($data === null) {

                // In both cases, we return an error message
                $response = [];
                $response["status"] = "ERROR";
                $response["message"] = "incorrectly formatted request";
                respond($response);

            } else {

                // If this worked, use the default key for the response from now on
                $crypt_key = $SETTINGS["WraithInitialCryptKey"];

            }

        } else {

            // If this worked, use the switch key for the response from now on
            $crypt_key = $SETTINGS["WraithSwitchCryptKey"];

        }

        // At this point, we should have a variable named $data holding
        // data received from the Wraith.

        // We just need to make sure that the data is an array and the required
        // headers are present

        // Check if the data is an array
        if (!(is_array($data))) {

            $response = [];
            $response["status"] = "ERROR";
            $response["message"] = "incorrectly formatted request";
            respond($response);

        }

        // Make sure the array has the required keys
        if (!(has_keys($data, [
            "req_type", // So we know what to do with the request
        ]))) {

            $response = [];
            $response["status"] = "ERROR";
            $response["message"] = "incorrectly formatted request";
            respond($response);

        }

        // The Wraith's request has now been fully validated and prepared and can
        // be processed

    } elseif ($requester === "panel") {

        // TODO


    } else {

        // This will never happen if the code is unmodified. However, to gracefully
        // handle mistakes in modification, this should stay here
        $response = [];
        $response["status"] = "ERROR";
        $response["message"] = "the request was identified but methods for handling it were not implemented";
        respond($response);

    }

    // Unset everything other than the required variables to save resources and namespace
    $keep_variables = [
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
        "requester_IP",
        "protocol_version",
        "crypt",
        "crypt_key",
        // This variable
        "keep_variables"
    ];

    foreach (get_defined_vars() as $name => $value) {

        if (!(in_array($name, $keep_variables))) {

            unset($$name);

        }

    }
    unset($keep_variables, $name, $value);

    /*

    REQUEST PROCESSING

    */

    // Create an instance of the handler class for the specified protocol
    $handler_class_name = "Handler_proto_v_".$protocol_version;
    $handler = new $handler_class_name($db, $requester, $requester_IP, $data, $SETTINGS);

    // Handle the request using the created handler
    $handler->handle_request();

    // Unset (destroy) the handler so it can clean up and respond to the
    // requester
    unset($handler);

    // If nothing was sent until now, assume that there was some
    // kind of error, most likely due to a faulty protocol handler
    $response = [];
    $response["status"] = "ERROR";
    $response["message"] = "no response generated";
    respond($response);

} else {

    http_response_code(405);
    header("Content-Type: text/plain");
    header("Allow: GET, POST");
    die("Unsupported method");

}
