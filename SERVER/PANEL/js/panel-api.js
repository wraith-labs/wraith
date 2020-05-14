// This is a PHP file returning a JavaScript function with the nescessary
// placeholders filled in

// A function which sends requests to the API and returns a processed response
function api(data) {

    // Define the required constants
    const request = new XMLHttpRequest();
    const APIPrefix = window.wraithManagerStorage["config"]["APIPrefix"];
    const firstLayerEncryptionKey = window.wraithManagerStorage["config"]["managementFirstLayerEncryptionKey"];
    const sessionID = window.wraithManagerStorage["config"]["sessionID"];
    const sessionToken = window.wraithManagerStorage["config"]["sessionToken"];
    const APILocation = window.wraithManagerStorage["config"]["APILocation"];
    const trustedAPIFingerprint = window.wraithManagerStorage["config"]["APIFingerprint"];
    const protocolVersion = 0;

    // Generate the full prefix for the message
    var fullPrefix = APIPrefix; // Start with API prefix
    fullPrefix += ["2", "4", "6", "8"][Math.floor(Math.random() * 4)]; // Add panel ID char
    fullPrefix += protocolVersion;

    // JSON encode the data to be sent
    data = JSON.stringify(data);
    // and encrypt it with the session token (second layer)
    var finalData = aes.encrypt(data, sessionToken);

    // Generate the first layer
    var firstLayerPayload = [
        sessionID,
        finalData,
    ];
    // JSON encode the first layer
    firstLayerPayload = JSON.stringify(firstLayerPayload);
    // Encrypt the first layer with the first layer key
    const finalPayload = aes.encrypt(firstLayerPayload, firstLayerEncryptionKey);

    // Define what happens when a response from the API is received
    var finalResponse;
    request.onreadystatechange = function() {
        // If the response was received successfully
        if (this.readyState == 4 && this.status == 200) {
            console.log(this.responseText);
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
