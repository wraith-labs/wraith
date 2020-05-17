<?php

// SQLite database file location relative to API.php
$DATABASE_LOCATION = "./storage/wraithdb";

// Get a database instance
// This can be edited to use MySQL or equivalent databases. As long as
// there is a $db variable holding a PDO database connection
// all should work (untested).
$GLOBALS["db"] = new PDO("sqlite:" . $DATABASE_LOCATION);
$db->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

// Check whether the database is initialised
try {

    $db->query("SELECT * FROM DB_INIT_INDICATOR")->fetchAll();

} catch (PDOException $e) {

    // If not, prepare the database

    // SQL Commands to be executed to initialise the database
    $dbInitCommands = [
        // SETTINGS Table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Settings` (
            `key` TEXT NOT NULL UNIQUE PRIMARY KEY,
            `value` TEXT
        );",
        // STATS Table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Stats` (
            `key` TEXT NOT NULL UNIQUE PRIMARY KEY,
            `value` TEXT
        );",
        // EVENTS Table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_EventHistory` (
            `eventID` TEXT NOT NULL UNIQUE PRIMARY KEY,
            `eventType` TEXT,
            `eventTime` TEXT,
            `eventProperties` TEXT
        );",
        // CONNECTED WRAITHS Table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_ActiveWraiths` (
            `assignedID` TEXT NOT NULL UNIQUE PRIMARY KEY,
            `hostProperties` TEXT,
            `wraithProperties` TEXT,
            `lastHeartbeatTime` TEXT,
            `issuedCommands` TEXT
        );",
        // COMMAND QUEUE Table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_CommandsIssued` (
            `commandID` TEXT NOT NULL UNIQUE PRIMARY KEY,
            `commandName` TEXT,
            `commandParams` TEXT,
            `commandTargets` TEXT,
            `commandResponses` TEXT,
            `timeIssued` TEXT
        );",
        // USERS Table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Users` (
            `userName` TEXT NOT NULL UNIQUE PRIMARY KEY,
            `userPassword` TEXT,
            `userPrivileges` TEXT,
            `userFailedLogins` INTEGER,
            `userFailedLoginsTimeoutStart` TEXT
        );",
        // SESSIONS Table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Sessions` (
            `sessionID` TEXT NOT NULL UNIQUE PRIMARY KEY,
            `username` TEXT,
            `sessionToken` TEXT,
            `lastSessionHeartbeat` TEXT
        );",
        // SETTINGS entries
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'wraithMarkOfflineDelay',
            '16'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'wraithInitialCryptKey',
            'QWERTYUIOPASDFGHJKLZXCVBNM'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'wraithSwitchCryptKey',
            'QWERTYUIOPASDFGHJKLZXCVBNM'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'APIFingerprint',
            'ABCDEFGHIJKLMNOP'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'wraithDefaultCommands',
            '" . json_encode([]) . "'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'APIPrefix',
            'W_'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'requestIPBlacklist',
            '" . json_encode([]) . "'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'managementSessionExpiryDelay',
            '12'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'managementFirstLayerEncryptionKey',
            '" . bin2hex(random_bytes(25)) . "'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'managementIPWhitelist',
            '[]'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'managementBruteForceMaxAttempts',
            '3'
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'managementBruteForceTimeoutSeconds',
            '300'
        );",
        // STATS Entries
        "INSERT INTO `WraithAPI_Stats` VALUES (
            'databaseSetupAPIVersion',
            '" . constant("API_VERSION") . "'
        );",
        "INSERT INTO `WraithAPI_Stats` VALUES (
            'databaseSetupTime',
            '" . time() . "'
        );",
        "INSERT INTO `WraithAPI_Stats` VALUES (
            'totalWraithConnections',
            '0'
        );",
        "INSERT INTO `WraithAPI_Stats` VALUES (
            'totalCommandsIssued',
            '0'
        );",
        "INSERT INTO `WraithAPI_Stats` VALUES (
            'totalManagerLogins',
            '0'
        );",
        // Mark the database as initialised
        "CREATE TABLE IF NOT EXISTS `DB_INIT_INDICATOR` (
            `DB_INIT_INDICATOR` INTEGER
        );"
    ];

    // Execute the SQL queries to initialise the database
    foreach ($dbInitCommands as $command) {

        $db->exec($command);

    }

}

// Set the global SETTINGS variable with the settings from the database
$settingsTable = $db->query("SELECT * FROM WraithAPI_Settings");
$SETTINGS = [];
foreach ($settingsTable as $tableRow) {
    $SETTINGS[$tableRow[0]] = $tableRow[1];
}

// Check whether a user account exists
// There has to be a way to manage the API so if there are no users,
// create one.
try {

    $API_USERS = $db->query("SELECT * FROM WraithAPI_Users")->fetchAll();

    if (sizeof($API_USERS) == 0) {
        throw new Exception("");
    }

} catch (Exception $e) {

    // Create default super admin user

    $userCreationCommand = "INSERT INTO `WraithAPI_Users` (
        `userName`,
        `userPassword`,
        `userPrivileges`
    ) VALUES (
        'SuperAdmin',
        '" . password_hash("SuperAdminPassword", PASSWORD_BCRYPT) . "',
        '2'
    );";

    $db->exec($userCreationCommand);

}

// Set the global API_USERS variable
$API_USERS = $db->query("SELECT * FROM WraithAPI_Users")->fetchAll();

// Functions for database management

// WRAITH

// Add a Wraith to the database
function dbAddWraith($wraith) {

    global $SETTINGS, $db;

    $statement = $db->prepare("INSERT INTO `WraithAPI_ActiveWraiths` (
        `assignedID`,
        `hostProperties`,
        `wraithProperties`,
        `lastHeartbeatTime`,
        `issuedCommands`
    ) VALUES (
        :assignedID,
        :hostProperties,
        :wraithProperties,
        :lastHeartbeatTime,
        :issuedCommands
    )");

    // Bind the parameters in the query with variables
    $statement->bindParam(":assignedID", $wraith["assignedID"]);
    $statement->bindParam(":hostProperties", $wraith["hostProperties"]);
    $statement->bindParam(":wraithProperties", $wraith["wraithProperties"]);
    $statement->bindParam(":lastHeartbeatTime", $wraith["lastHeartbeatTime"]);
    $statement->bindParam(":issuedCommands", $wraith["issuedCommands"]);

    // Execute the statement to add a Wraith
    $statement->execute();

}

// Remove Wraith(s)
function dbRemoveWraiths($ids) {

    global $SETTINGS, $db;

    $statement = $db->prepare("DELETE FROM `WraithAPI_ActiveWraiths` WHERE assignedID == :IDToDelete");

    // Remove each ID
    foreach ($ids as $id) {
        $statement->bindParam(":IDToDelete", $id);
        $statement->execute();
    }

}

// Check which Wraiths have not sent a heartbeat in the mark dead time and remove
// them from the database
function dbExpireWraiths() {

    global $SETTINGS, $db;

    // Remove all Wraith entries where the last heartbeat time is older than
    // the $SETTINGS["wraithMarkOfflineDelay"]
    $statement = $db->prepare("DELETE FROM `WraithAPI_ActiveWraiths`
        WHERE `lastHeartbeatTime` < :earliestValidHeartbeat");

    // Get the unix timestamp for $SETTINGS["wraithMarkOfflineDelay"] seconds ago
    $earliestValidHeartbeat = time()-$SETTINGS["wraithMarkOfflineDelay"];
    $statement->bindParam(":earliestValidHeartbeat", $earliestValidHeartbeat);

    // Execute
    $statement->execute();

}

// Get a list of Wraiths and their properties
function dbGetWraiths() {

    global $SETTINGS, $db;

    // Get a list of wraiths from the database
    $wraiths_db = $db->query("SELECT * FROM WraithAPI_ActiveWraiths")->fetchAll();

    // Create an array to store a processed list of wraiths
    $wraiths = [];

    foreach ($wraiths_db as $wraith) {

        // Move the assigned ID to a separate variable
        $wraithID = $wraith["assignedID"];
        unset($wraith["assignedID"]);

        // Set the (assignedID-less) wraith array as an element of the $wraiths
        // array
        $wraiths[$wraithID] = $wraith;

    }

    // Return the processed wraiths array
    return $wraiths;

}

// COMMAND

// Issue a command to Wraith(s)
function dbIssueCommand($command) {

    global $SETTINGS, $db;

    // TODO

}

// Delete command(s) from the command table
function dbCancelCommands($ids) {

    global $SETTINGS, $db;

    // TODO
    $statement = $db->prepare("DELETE FROM `WraithAPI_CommandsIssued` WHERE assignedID == :IDToDelete");

    // Remove each ID
    foreach ($ids as $id) {
        $statement->bindParam(":IDToDelete", $id);
        $statement->execute();
    }

}

// Get command(s)
function dbGetCommands($ids) {

    global $SETTINGS, $db;

    // TODO

}

// SETTINGS

// Edit an API setting
function dbSetSetting($setting, $value) {

    global $SETTINGS, $db;

    // Update setting value
    $statement = $db->prepare("UPDATE WraithAPI_Settings
        SET `value` = :value WHERE `key` = :setting;");

    // Bind the required parameters
    $statement->bindParam(":setting", $setting);
    $statement->bindParam(":value", $value);

    // Execute
    $statement->execute();

}

// Get the value of a setting
function dbGetSetting($setting) {

    global $SETTINGS, $db;

    // Get the setting directly from the database. This function should
    // only be used to get the most up-to-date values of settings as
    // using the $SETTINGS array is easier and more efficient
    $statement = $db->prepare("SELECT * FROM WraithAPI_Settings
        WHERE `key` = :setting LIMIT 1");

    // Bind parameters
    $statement->bindParam(":setting", $setting);

    // Execute
    $statement->execute();

    // Fetch results
    $value = $statement->fetchAll()[0]["value"];

    // Return results
    return $value;

}

// USERS

// Create a new user
function dbAddUser($user) {

    global $SETTINGS, $db;

    // TODO

}

// Change username
function dbChangeUserName($currentUsername, $newUsername) {

    global $SETTINGS, $db;

    // TODO

}

// Change user password
function dbChangeUserPass($username, $newPassword) {

    global $SETTINGS, $db;

    // TODO

}

// Change user privilege level (0=User, 1=Admin, 2=SuperAdmin)
function dbChangeUserPrivilege($username, $newPrivilegeLevel) {

    global $SETTINGS, $db;

    // TODO

}

// SESSIONS

// Create a session for a user
function dbCreateSession($username) {

    global $SETTINGS, $db;

    $statement = $db->prepare("INSERT INTO `WraithAPI_Sessions` (
        `sessionID`,
        `username`,
        `sessionToken`,
        `lastSessionHeartbeat`
    ) VALUES (
        :sessionID,
        :username,
        :sessionToken,
        :lastSessionHeartbeat
    )");

    // Create session variables
    $sessionID = uniqid();
    $sessionToken = bin2hex(random_bytes(25));
    $lastSessionHeartbeat = time();

    // Bind the parameters in the query with variables
    $statement->bindParam(":username", $username);
    $statement->bindParam(":sessionID", $sessionID);
    $statement->bindParam(":sessionToken", $sessionToken);
    $statement->bindParam(":lastSessionHeartbeat", $lastSessionHeartbeat);

    // Execute the statement to add a Wraith
    $statement->execute();

    // Return the ID of the created session
    return $sessionID;

}

// Get a list of all sessions
function dbGetSessions() {

    global $SETTINGS, $db;

    // Get a list of sessions from the database
    $sessions_db = $db->query("SELECT * FROM WraithAPI_Sessions")->fetchAll();

    // Create an array to store a processed list of sessions
    $sessions = [];

    foreach ($sessions_db as $session) {

        // Move the session ID to a separate variable
        $sessionID = $session["sessionID"];
        unset($session["sessionID"]);

        // Set the (sessionID-less) session array as an element of the $sessions
        // array
        $sessions[$sessionID] = $session;

    }

    // Return the processed sessions array
    return $sessions;

}

// Delete a session
function dbDestroySession($sessionID) {

    global $SETTINGS, $db;

    // Remove the session with the specified ID
    $statement = $db->prepare("DELETE FROM `WraithAPI_Sessions`
        WHERE `sessionID` = :sessionID");

    // Bind parameters
    $statement->bindParam(":sessionID", $sessionID);

    // Execute
    $statement->execute();

}

// Delete sessions which have not had a heartbeat recently
function dbExpireSessions() {

    global $SETTINGS, $db;

    // Remove all sessions where the last heartbeat time is older than
    // the $SETTINGS["managementSessionExpiryDelay"]
    $statement = $db->prepare("DELETE FROM `WraithAPI_Sessions`
        WHERE `lastSessionHeartbeat` < :earliestValidHeartbeat");

    // Get the unix timestamp for $SETTINGS["managementSessionExpiryDelay"] seconds ago
    $earliestValidHeartbeat = time()-$SETTINGS["managementSessionExpiryDelay"];
    $statement->bindParam(":earliestValidHeartbeat", $earliestValidHeartbeat);

    // Execute
    $statement->execute();

}

// Update the session last heartbeat time
function dbUpdateSessionLastHeartbeat($sessionID) {

    global $SETTINGS, $db;

    // Update the last heartbeat time to the current time
    $statement = $db->prepare("UPDATE WraithAPI_Sessions
        SET `lastSessionHeartbeat` = :currentTime WHERE `sessionID` = :sessionID;");

    // Get the current time and the session ID and bind them to the params
    $statement->bindParam(":currentTime", time());
    $statement->bindParam(":sessionID", $sessionID);

    // Execute
    $statement->execute();

}

// STATS

// Update a statistic
function dbGetStats() {

    global $SETTINGS, $db;

    // Get a list of statistics from the database
    $stats_db = $db->query("SELECT * FROM WraithAPI_Stats")->fetchAll();

    // Create an array to store a processed list of statistics
    $stats = [];

    foreach ($stats_db as $stat) {

        // Get the stat key
        $key = $stat["key"];

        // Copy the stat value to the stats array
        $stats[$key] = $stat["value"];

    }

    // Return the processed stats array
    return $stats;

}

// Update a statistic
function dbUpdateStat($stat, $value) {

    global $SETTINGS, $db;

    // Update a stat
    $statement = $db->prepare("UPDATE WraithAPI_Stats
        SET `value` = :value WHERE `key` = :stat;");

    // Bind the parameters
    $statement->bindParam(":stat", $stat);
    $statement->bindParam(":value", $value);

    // Execute
    $statement->execute();

}

// MISC

// Re-generate the first-layer encryption key for management sessions
function dbRegenMgmtCryptKeyIfNoSessions() {

    global $SETTINGS, $db;

    // If there are no active sessions
    $allSessions = dbGetSessions();
    if (sizeof($allSessions) == 0) {

        // Update the first layer encryption key
        dbSetSetting("managementFirstLayerEncryptionKey", bin2hex(random_bytes(25)));

    }

}
