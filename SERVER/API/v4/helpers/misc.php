<?php

// Get IP of client
function getClientIP($strict = false, $fallback = null) {

    if ($fallback) {

        $IPAddress = $fallback;

    } else {

        $IPAddress = "UNKNOWN";

    }

    if ($strict) {

        $keys = ['REMOTE_ADDR'];

    } else {

        $keys = ['HTTP_CLIENT_IP', 'HTTP_X_FORWARDED_FOR', 'HTTP_X_FORWARDED', 'HTTP_FORWARDED_FOR', 'HTTP_FORWARDED', 'REMOTE_ADDR'];

    }

    foreach($keys as $k) {

        if (isset($_SERVER[$k]) && !empty($_SERVER[$k]) && filter_var($_SERVER[$k], FILTER_VALIDATE_IP)) {

            $IPAddress = $_SERVER[$k];
            break;

        }

    }

    return $IPAddress;

}

// Check if an array has all of the keys
function hasKeys($array, $keys) {

    if (!(count(array_diff($keys, array_keys($array))) === 0)) {

        return false;

    } else {

        return true;

    }

}

// Generate a fingerprint for a Wraith based on an array of data
// likely to be unique to that Wraith
function genWraithFingerprint($fingerprintableData) {

    // Join the fingerprintable data array with a delimeter to limit overlap of fingerprints
    // Without a delimeter, the arrays ["abc", "def"] and ["abcd", "ef"] would produce the same
    // fingerprint
    $fingerprintableData = join("|||", $fingerprintableData);

    // Hash the resulting string to generate the fingerprint
    $fingerprint = hash("md5", $fingerprintableData);

    return $fingerprint;

}