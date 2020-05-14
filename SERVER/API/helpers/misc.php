<?php

// Get IP of client
function getClientIP() {

    $IPAddress = 'UNKNOWN';
    $keys = array('HTTP_CLIENT_IP','HTTP_X_FORWARDED_FOR','HTTP_X_FORWARDED','HTTP_FORWARDED_FOR','HTTP_FORWARDED','REMOTE_ADDR');
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
