// This is a PHP file returning a JavaScript function with the nescessary
// placeholders filled in

// A function which sends requests to the API and returns a processed response
async function api(data) {

    // Define the required constants
    const request = new XMLHttpRequest();
    const APIPrefix = window.wraithManagerStorage["config"]["APIPrefix"];
    const firstLayerEncryptionKey = window.wraithManagerStorage["config"]["firstLayerEncryptionKey"];
    const sessionID = window.wraithManagerStorage["config"]["sessionID"];
    const sessionToken = window.wraithManagerStorage["config"]["sessionToken"];
    const APILocation = window.wraithManagerStorage["config"]["APILocation"];
    const trustedAPIFingerprint = window.wraithManagerStorage["config"]["APIFingerprint"];
    const protocolVersion = 0;

    // Attach the session token to the data
    data["sessionToken"] = sessionToken;

    // Generate the full prefix for the message
    var fullPrefix = APIPrefix; // Start with API prefix
    fullPrefix += ["2", "4", "6", "8"][Math.floor(Math.random() * 4)]; // Add panel ID char
    fullPrefix += protocolVersion;

    // Generate the second layer
    // JSON encode the data to be sent
    var finalData = JSON.stringify(data);
    // and encrypt it with the session token (second layer)
    finalData = aes.encrypt(finalData, sessionToken);

    // Generate the first layer
    finalData = [
        sessionID,
        finalData,
    ];
    // JSON encode the first layer
    finalData = JSON.stringify(finalData);
    // Encrypt the first layer with the first layer key and add the prefix
    finalData = fullPrefix + aes.encrypt(finalData, firstLayerEncryptionKey);

    // Define what happens when a response from the API is received
    var finalResponse;
    request.onreadystatechange = function() {

        // If the response was received successfully
        if (this.readyState == 4 && this.status == 200) {

            // Create a variable to hold the API response
            var response;

            // The response could be encrypted but it might not be
            try {

                // Set the response to the data from the API
                response = request.responseText;

                // Check if the response starts with the prefix.
                if (!(response.startsWith(APIPrefix))) {

                    // If it does not, it's either not encrypted or
                    // invalid so skip straight to the unencrypted
                    // parsing
                    throw "No Prefix";

                }

                // Remove the prefix
                response = response.replace(APIPrefix, "");

                // Try decrypting the response (API responses are only encrypted)
                // with the session token
                response = aes.decrypt(response, sessionToken);
                // Parse the JSON response
                response = JSON.parse(response);
                // Verify the fingerprint of the server (only if the communication is encrypted because otherwise the server does not attach it)
                if (response["APIFingerprint"] != trustedAPIFingerprint) {

                    console.log("WARNING: The API did not provide a trusted fingerprint (expected = `" + response_dict["server_id"] + "` | received = `" + trusted_server_signature + "`. The API response will be ignored.");
                    return;

                }

                // If we got here, the response was processed successfully so
                // can be assigned to finalResponse
                finalResponse = response;

            } catch (err) {

                // If we got here, the response was not encrypted or invalid

                // Re-set the response variable
                response = request.responseText;

                // Try to JSON parse the response
                try {

                    response = JSON.parse(response);

                // If this fails, log the error and return
                } catch {

                    console.log("ERROR: The API response could not be parsed.");
                    return;

                }

                // If we got here, the response was not encrypted but processed
                // successfully so can be assigned to finalResponse
                finalResponse = response;

            }

            // If the API does not give a status
            if (!("status" in finalResponse)) {

                // Assume success
                finalResponse["status"] = "SUCCESS";

            }

            // Check the status reported by the API
            if (finalResponse["status"] == "SUCCESS") {

                // Nothing (currently)

            } else if (finalResponse["status"] == "ERROR") {

                // Output a warning to the console
                console.log("WARNING: The API returned an error.")

            } else {

                // No other status is valid
                console.log("ERROR: The API provided an invalid status code.")

            }

            // If the API returns a message, log it
            if ("message" in finalResponse) {

                console.log("Message from API: " + finalResponse["message"]);

            }

        // If the API returns a HTTP error code
        } else if (this.readyState == 4 && this.status != 200) {

            console.log("ERROR: The API returned a non-200 code when called (" + this.status + ").");
            // Set both response variables to undefined so this error can be detected down the line
            return;

        }

    }

    // Send the API request
    request.open("POST", APILocation, false);
    await request.send(finalData);

    // Return the response
    return finalResponse;

}
