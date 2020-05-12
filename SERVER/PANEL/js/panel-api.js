// This is a PHP file returning a JavaScript function with the nescessary
// placeholders filled in

// A function which sends requests to the API and returns a processed response
function api(data, token=null) {

    // If a token is specified, add it to the request
    if (token != null) {
        data["managementAuthCode"] = token;
    }

    // Define the required constants
    const request = new XMLHttpRequest();

    // TODO: Make these variable
    const APIPrefix = "W_";
    const APIManagementAuthToken = "";
    const APILocation = "http://localhost/API/api.php";
    const cryptKey = "QWERTY";
    const protocolVersion = 0;
    const trustedAPIFingerprint = "ABCDEFGHIJKLMNOP";

    // Generate the full prefix for the message
    var fullPrefix = APIPrefix; // Start with API prefix
    fullPrefix += ["2", "4", "6", "8"][Math.floor(Math.random() * 4)]; // Add panel ID char
    fullPrefix += protocolVersion;

    // JSON encode the data to be sent
    data = JSON.stringify(data);
    // and encrypt it + add the prefix
    var finalData = fullPrefix + aes.encrypt(data, cryptKey);

    // Define what happens when a response from the API is received
    var finalResponse;
    request.onreadystatechange = function() {
        // If the response was received successfully
        if (this.readyState == 4 && this.status == 200) {
            // The response could be encrypted but it might not be
            try {
                // Try decrypting the response
                response = aes.decrypt(request.responseText, cryptKey);
                // Parse the JSON response
                finalResponse = JSON.parse(response);
                // Verify the fingerprint of the server (only if the communication is encrypted because otherwise the server does not attach it)
                if (finalResponse["APIFingerprint"] != trustedAPIFingerprint) {
                    throw new Error("the API did not provide a trusted fingerprint (expected = `" + response_dict["server_id"] + "` | received = `" + trusted_server_signature + "`");
                }
            } catch (err) {
                // If we got here, the response was not encrypted or the server provided the wrong ID
                // If the latter, re-throw the error so the user can see
                if (err.message.startsWith("The server provided an incorrect ID. ")) { throw err; }
                // If we got here, that means it was just a simple decryption problem so get the response as plaintext
                response = request.responseText;
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
            console.log("the API returned a non-200 code when called (" + this.status + ")");
            // Set both response variables to undefined so this error can be detected down the line
            response = undefined;
            response_dict = undefined;
        }
    }


}

/*


    // Define the required constants
    const request = new XMLHttpRequest();
    const apiPrefix = "W_"; // TODO: Get this from somewhere
    const apiURL = "/API/api.php"; // TODO: use actual API URL
    const key = current_panel_crypt_key;
    const data_json = JSON.stringify(data);

    // Generate a prefix by which the API can identify that the request is from a valid source
    const crypt_prefix_charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890";
    var crypt_prefix = "";
    for (var i = 0; i < 5; i++) {
        crypt_prefix += crypt_prefix_charset[Math.floor(Math.random() * crypt_prefix_charset.length)];
    }
    crypt_prefix += "Wr7H"

    // Define the variables to hold the response from the API
    var response = "";
    var response_dict = {};

    // When a response is received
    req.onreadystatechange = function() {
        // If the response was received successfully
        if (this.readyState == 4 && this.status == 200) {
            // The response could be encrypted but it might not be
            try {
                // Try decrypting the response
                response = aes.decrypt(request.responseText, key);
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
                response = request.responseText;
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
            console.log("the API returned a non-200 code when called (" + this.status + ")");
            // Set both response variables to undefined so this error can be detected down the line
            response = undefined;
            response_dict = undefined;
        }
    }

    // Actually send the request
    request.open("POST", apiURL, true);
    request.send(crypt_prefix+aes.encrypt(args_json, key));

    // Return the response from the API to the caller
    return response_dict;
} */
