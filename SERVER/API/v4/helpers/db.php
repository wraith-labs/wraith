<?php

// TODO - add event management methods and thoroughly test every method

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

        // Set database error handling policy
        $this->db->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

        // Start a transaction (prevent modification to the database by other
        // scripts running at the same time). If a transaction is currently in
        // progress, this will error so a try/catch and a loop is needed.
        while (true) {

            try {

                $this->db->beginTransaction();
                break;

            } catch (PDOException $e) {}

        }

        // Define the SQL commands used to initialise the database
        $this->dbInitCommands = [

            // CONNECTED WRAITHS Table
            "CREATE TABLE IF NOT EXISTS `WraithAPI_ActiveWraiths` (
                `assignedID` TEXT NOT NULL UNIQUE PRIMARY KEY,
                `hostProperties` TEXT,
                `wraithProperties` TEXT,
                `lastHeartbeatTime` TEXT,
                `issuedCommands` TEXT
            );",
            // COMMAND QUEUE Table
            "CREATE TABLE IF NOT EXISTS `WraithAPI_IssuedCommands` (
                `assignedID` TEXT NOT NULL UNIQUE PRIMARY KEY,
                `commandName` TEXT,
                `commandParams` TEXT,
                `commandTargets` TEXT,
                `timeIssued` TEXT
            );",
            // SETTINGS Table
            "CREATE TABLE IF NOT EXISTS `WraithAPI_Settings` (
                `key` TEXT NOT NULL UNIQUE PRIMARY KEY,
                `value` TEXT
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
                `assignedID` TEXT NOT NULL UNIQUE PRIMARY KEY,
                `username` TEXT,
                `creatorIP` TEXT,
                `sessionToken` TEXT,
                `lastHeartbeatTime` TEXT
            );",
            // EVENTS Table
            "CREATE TABLE IF NOT EXISTS `WraithAPI_EventHistory` (
                `assignedID` TEXT NOT NULL UNIQUE PRIMARY KEY,
                `eventType` TEXT,
                `eventTargets` TEXT,
                `eventTime` TEXT,
                `eventData` TEXT
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
                '" . bin2hex(random_bytes(25)) . "'
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

        }

        // Check if a user account exists (LIMIT 1 for efficiency)
        if (sizeof($this->dbGetUsers([], 1, 0)) < 1) {

            // A user should be added to allow managing the API
            $this->dbAddUser([
                // This is the only place referencing the default account
                // SuperAdmin:SuperAdminPass so the default credentials can
                // easily be changed if needed.
                "userName" => "SuperAdmin",
                "userPassword" => "SuperAdminPass",
                "userPrivilegeLevel" => 2
            ]);

        }

    }

    // On object destruction
    function __destruct() {

        // Commit database changes (write changes made during the runtime of the
        // script to the database and allow other scripts to access the database)
        $this->db->commit();

        // Close the database connection
        $this->db = null;

    }

    // HELPERS (internal)

    // Execute SQL on the database with optional parameters using secure
    // prepared statements
    private function SQLExec($SQL, $params = []) {

        $statement = $this->db->prepare($SQL);

        $statement->execute($params);

        // Return the statement so further actions can be performed on it like
        // fetchAll().
        return $statement;

    }

    // Convert an array into a SQL WHERE clause for use as a filter
    // Adapted from https://stackoverflow.com/a/62181134/8623347
    private function generateFilter($filter, $columnNameWhitelist, $limit = -1, $offset = -1) {

        $conditions = [];
        $parameters = [];

        foreach ($filter as $key => $values) {

            // Ensure that the column names are whitelisted to prevent SQL
            // injection
            if (array_search($key, $columnNameWhitelist, true) === false) {

                throw new InvalidArgumentException("invalid field name in filter");

            }

            // Generate the SQL for each condition and add the values to the list
            // of parameters
            $conditions[] = "`$key` in (".str_repeat('?,', count($values) - 1) . '?'.")";
            $parameters = array_merge($parameters, $values);

        }

        // Generate the SQL (no SQL needs to be generated if no conditions
        // were given)
        $sql = "";
        if ($conditions) {

            $sql .= " WHERE " . implode(" AND ", $conditions);

        }

        // Add the LIMIT and OFFSET for pagination

        if ((int)$limit >= 0) {

            $sql .= " LIMIT " . (int)$limit;

        }

        if ((int)$offset >= 1) {

            $sql .= " OFFSET " . (int)$offset;

        }

        // The filter should now be translated into valid SQL and parameters
        // so it can be returned
        return [$sql, $parameters];

    }

    // DATABASE MANAGEMENT (internal)

    // Check if the database has been initialised
    private function isDatabasePostInit() {

        // Check if the DB_INIT_INDICATOR table exists
        $statement = $this->SQLExec("SELECT name FROM sqlite_master WHERE type='table' AND name='DB_INIT_INDICATOR' LIMIT 1");

        // Convert the result into a boolean
        // The result will be an array of all tables named "DB_INIT_INDICATOR"
        // If the array is of length 0 (no such table), the boolean will be false.
        // All other cases result in true (the only other possible case here is 1).
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

                $this->SQLExec($command);

            } catch (PDOException $e) {

                return false;

            }

        }

        // Create a database initialisation event
        $this->dbAddEvent([
            "eventType" => "initOfDB",
            "eventTargets" => ["API"],
            "eventData" => [
                "description" => "Database initialised"
            ]
        ]);

        // If false was not yet returned, everything was successful
        return true;

    }

    // Delete all Wraith API tables from the database
    // (init will not be called automatically)
    private function clearDB() {

        // The following will generate an array of SQL commands which will
        // delete every table in the database
        $statement = $this->SQLExec("SELECT 'DROP TABLE ' || name ||';' FROM sqlite_master WHERE type = 'table'");

        // Get the SQL commands
        $commands = $statement->fetchAll();

        foreach ($commands as $command) {

            $this->SQLExec($command[0]);

        }

    }

    // ACTIVE WRAITH TABLE MANAGEMENT (public)

    // Add a Wraith to the database
    function dbAddWraith($data) {

        // Check parameters and set defaults
        if (!(array_key_exists("assignedID", $data))) {

            $data["assignedID"] = uniqid();

        }
        if (!(array_key_exists("hostProperties", $data))) {

            // The hostProperties have no default value and are required
            return false;

        }
        if (!(array_key_exists("wraithProperties", $data))) {

            // The wraithProperties have no default value and are required
            return false;

        }
        if (!(array_key_exists("lastHeartbeatTime", $data))) {

            $data["lastHeartbeatTime"] = time();

        }
        if (!(array_key_exists("issuedCommands", $data))) {

            $data["issuedCommands"] = [];

        }

        $SQL = "INSERT INTO `WraithAPI_ActiveWraiths` (
                `assignedID`,
                `hostProperties`,
                `wraithProperties`,
                `lastHeartbeatTime`,
                `issuedCommands`
            ) VALUES (
                ?,
                ?,
                ?,
                ?,
                ?
            )";

        $params = [
            $data["assignedID"],
            json_encode($data["hostProperties"]),
            json_encode($data["wraithProperties"]),
            $data["lastHeartbeatTime"],
            json_encode($data["issuedCommands"])
        ];

        $this->SQLExec($SQL, $params);

        return $data["assignedID"];

    }

    // Remove Wraith(s)
    function dbRemoveWraiths($filter = [], $limit = -1, $offset = -1) {

        $validFilterColumnNames = [
            "assignedID",
            "hostProperties",
            "wraithProperties",
            "lastHeartbeatTime",
            "issuedCommands"
        ];

        $SQL = "DELETE FROM `WraithAPI_ActiveWraiths`";

        $params = [];

        // Apply the filters
        $filterSQL = $this->generateFilter($filter, $validFilterColumnNames, $limit, $offset);
        $SQL .= $filterSQL[0];
        $params = array_merge($params, $filterSQL[1]);

        $statement = $this->SQLExec($SQL, $params);

    }

    // Get a list of Wraiths and their properties
    function dbGetWraiths($filter = [], $limit = -1, $offset = -1) {

        $validFilterColumnNames = [
            "assignedID",
            "hostProperties",
            "wraithProperties",
            "lastHeartbeatTime",
            "issuedCommands"
        ];

        $SQL = "SELECT * FROM WraithAPI_ActiveWraiths";

        $params = [];

        // Apply the filters
        $filterSQL = $this->generateFilter($filter, $validFilterColumnNames, $limit, $offset);
        $SQL .= $filterSQL[0];
        $params = array_merge($params, $filterSQL[1]);

        $statement = $this->SQLExec($SQL, $params);

        // Get a list of wraiths from the database
        $wraithsDB = $statement->fetchAll();

        $wraiths = [];

        foreach ($wraithsDB as $wraith) {

            // Move the assigned ID to a separate variable
            $wraithID = $wraith["assignedID"];
            unset($wraith["assignedID"]);

            $wraiths[$wraithID] = $wraith;

        }

        return $wraiths;

    }

    // Update the Wraith last heartbeat time
    function dbUpdateWraithLastHeartbeat($assignedID, $timeToSet = null) {

        // Set $timeToSet to the current time if no value was passed
        $timeToSet = isset($timeToSet) ? $timeToSet : time();

        // Update the last heartbeat time to the current time
        $SQL = "UPDATE WraithAPI_ActiveWraiths SET `lastHeartbeatTime` = ? WHERE `assignedID` = ?";

        $params = [
            $timeToSet,
            $assignedID
        ];

        $this->SQLExec($SQL, $params);

    }

    // Check which Wraiths have not sent a heartbeat in the mark dead time and remove
    // them from the database
    function dbExpireWraiths() {

        // Remove all Wraith entries where the last heartbeat time is older than
        // the $SETTINGS["wraithMarkOfflineDelay"]
        $SQL = "DELETE FROM `WraithAPI_ActiveWraiths` WHERE `lastHeartbeatTime` < ?";

        $params = [
            // Get the unix timestamp for $SETTINGS["wraithMarkOfflineDelay"] seconds ago
            time()-$this->dbGetSettings(["key" => ["wraithMarkOfflineDelay"]])["wraithMarkOfflineDelay"]
        ];

        $this->SQLExec($SQL, $params);

    }

    // ISSUED COMMAND TABLE MANAGEMENT (public)

    // Issue a command to Wraith(s)
    function dbAddCommand($data) {

        // Check parameters and set defaults
        if (!(array_key_exists("assignedID", $data))) {

            $data["assignedID"] = uniqid();

        }
        if (!(array_key_exists("commandName", $data))) {

            // The commandName has no default value and is required
            return false;

        }
        if (!(array_key_exists("commandParams", $data))) {

            $data["commandParams"] = "";

        }
        if (!(array_key_exists("commandTargets", $data))) {

            // The commandTargets parameter has no default value and is required
            return false;

        }
        if (!(array_key_exists("timeIssued", $data))) {

            $data["timeIssued"] = time();

        }

        $SQL = "INSERT INTO `WraithAPI_IssuedCommands` (
                `assignedID`,
                `commandName`,
                `commandParams`,
                `commandTargets`,
                `timeIssued`
            ) VALUES (
                ?,
                ?,
                ?,
                ?,
                ?
            )";

        $params = [
            $data["assignedID"],
            $data["commandName"],
            $data["commandParams"],
            $data["commandTargets"],
            $data["timeIssued"],
        ];

        $this->SQLExec($SQL, $params);

        return $data["assignedID"];

    }

    // Delete command(s) from the command table
    function dbRemoveCommands($filter = [], $limit = -1, $offset = -1) {

        $validFilterColumnNames = [
            "assignedID",
            "commandName",
            "commandParams",
            "commandTargets",
            "timeIssued"
        ];

        $SQL = "DELETE FROM `WraithAPI_IssuedCommands`";

        $params = [];

        // Apply the filters
        $filterSQL = $this->generateFilter($filter, $validFilterColumnNames, $limit, $offset);
        $SQL .= $filterSQL[0];
        $params = array_merge($params, $filterSQL[1]);

        $statement = $this->SQLExec($SQL, $params);

    }

    // Get command(s)
    function dbGetCommands($filter = [], $limit = -1, $offset = -1) {

        $validFilterColumnNames = [
            "assignedID",
            "commandName",
            "commandParams",
            "commandTargets",
            "timeIssued"
        ];

        $SQL = "SELECT * FROM WraithAPI_IssuedCommands";

        $params = [];

        // Apply the filters
        $filterSQL = $this->generateFilter($filter, $validFilterColumnNames, $limit, $offset);
        $SQL .= $filterSQL[0];
        $params = array_merge($params, $filterSQL[1]);

        $statement = $this->SQLExec($SQL, $params);

        $eventsDB = $statement->fetchAll();

        $events = [];

        foreach ($eventsDB as $event) {

            // Move the assignedID to a separate variable
            $eventID = $event["assignedID"];
            unset($event["assignedID"]);

            $events[$eventID] = $event;

        }

        return $events;

    }

    // Get all commands for a Wraith
    function dbGetCommandsForWraith() {

        // TODO

    }

    // SETTINGS TABLE MANAGEMENT (public)

    // Edit an API setting
    function dbSetSetting($name, $value) {

        // Update setting value
        $SQL = "UPDATE WraithAPI_Settings SET `value` = ? WHERE `key` = ?";

        $params = [
            $value,
            $name
        ];

        $this->SQLExec($SQL, $params);

    }

    // Refresh the settings property of the DBManager
    function dbGetSettings($filter = [], $limit = -1, $offset = -1) {

        $validFilterColumnNames = [
            "key",
            "value"
        ];

        $SQL = "SELECT * FROM WraithAPI_Settings";

        $params = [];

        // Apply the filters
        $filterSQL = $this->generateFilter($filter, $validFilterColumnNames, $limit, $offset);
        $SQL .= $filterSQL[0];
        $params = array_merge($params, $filterSQL[1]);

        $statement = $this->SQLExec($SQL, $params);

        $result = $statement->fetchAll();

        // Format the results
        $settings = [];
        foreach ($result as $tableRow) {

            $settings[$tableRow[0]] = $tableRow[1];

        }

        return $settings;

    }

    // USERS TABLE MANAGEMENT (public)

    // Create a new user
    function dbAddUser($data) {

        // Check parameters and set defaults
        if (!(array_key_exists("userName", $data))) {

            // The userName has no default value and is required
            return false;

        }
        if (!(array_key_exists("userPassword", $data))) {

            // The userPassword has no default value and is required
            return false;

        }
        if (!(array_key_exists("userPrivilegeLevel", $data))) {

            $data["userPrivilegeLevel"] = 0;

        }

        // Hash the password
        $data["userPassword"] = password_hash($data["userPassword"], PASSWORD_BCRYPT);

        $this->SQLExec("INSERT INTO `WraithAPI_Users` (
                `userName`,
                `userPassword`,
                `userPrivileges`,
                `userFailedLogins`,
                `userFailedLoginsTimeoutStart`
            ) VALUES (
                ?,
                ?,
                ?,
                '0',
                '0'
            );",
            [
                $data["userName"],
                $data["userPassword"],
                $data["userPrivilegeLevel"]
            ]
        );

        return $data["userName"];

    }

    // Delete a user
    function dbRemoveUsers($filter = [], $limit = -1, $offset = -1) {

        $validFilterColumnNames = [
            "userName",
            "userPassword",
            "userPrivileges",
            "userFailedLoginsTimeoutStart"
        ];

        $SQL = "DELETE FROM `WraithAPI_Users`";

        $params = [];

        // Apply the filters
        $filterSQL = $this->generateFilter($filter, $validFilterColumnNames, $limit, $offset);
        $SQL .= $filterSQL[0];
        $params = array_merge($params, $filterSQL[1]);

        $statement = $this->SQLExec($SQL, $params);

    }

    // Get a list of users and their properties
    function dbGetUsers($filter = [], $limit = -1, $offset = -1) {

        $validFilterColumnNames = [
            "userName",
            "userPassword",
            "userPrivileges",
            "userFailedLogins",
            "userFailedLoginsTimeoutStart"
        ];

        $SQL = "SELECT * FROM WraithAPI_Users";

        $params = [];

        // Apply the filters
        $filterSQL = $this->generateFilter($filter, $validFilterColumnNames, $limit, $offset);
        $SQL .= $filterSQL[0];
        $params = array_merge($params, $filterSQL[1]);

        $statement = $this->SQLExec($SQL, $params);

        // Get a list of users from the database
        $usersDB = $statement->fetchAll();

        $users = [];

        foreach ($usersDB as $user) {

            // Move the userName to a separate variable
            $userName = $user["userName"];
            unset($user["userName"]);

            $users[$userName] = $user;

        }

        return $users;

    }

    // Change username
    function dbChangeUserName($currentUsername, $newUsername) {

        // Update userName value
        $SQL = "UPDATE WraithAPI_Users SET `userName` = ? WHERE `userName` = ?";

        $params = [
            $newUsername,
            $currentUsername
        ];

        $this->SQLExec($SQL, $params);

    }

    // Verify that a user password is correct
    function dbVerifyUserPass($username, $password) {

        $user = $this->dbGetUsers([
            "userName" => [$username]
        ])[$username];

        return password_verify($password, $user["userPassword"]);

    }

    // Change user password
    function dbChangeUserPass($username, $newPassword) {

        // Update userPassword value
        $SQL = "UPDATE WraithAPI_Users SET `userPassword` = ? WHERE `userName` = ?";

        $params = [
            password_hash($newPassword, PASSWORD_BCRYPT),
            $username
        ];

        $this->SQLExec($SQL, $params);

    }

    // Change user privilege level (0=User, 1=Admin, 2=SuperAdmin)
    function dbChangeUserPrivilege($username, $newPrivilegeLevel) {

        // Update userPassword value
        $SQL = "UPDATE WraithAPI_Users SET `userPrivileges` = ? WHERE `userName` = ?";

        $params = [
            $newPrivilegeLevel,
            $username
        ];

        $this->SQLExec($SQL, $params);

    }

    // SESSIONS TABLE MANAGEMENT (public)

    // Create a session for a user
    function dbAddSession($data) {

        // Check parameters and set defaults
        if (!(array_key_exists("assignedID", $data))) {

            $data["assignedID"] = uniqid();

        }
        if (!(array_key_exists("username", $data))) {

            // The username has no default value and is required
            return false;

        }
        if (!(array_key_exists("creatorIP", $data))) {

            $data["creatorIP"] = "*";

        }
        if (!(array_key_exists("sessionToken", $data))) {

            $data["sessionToken"] = bin2hex(random_bytes(25));

        }
        if (!(array_key_exists("lastHeartbeatTime", $data))) {

            $data["lastHeartbeatTime"] = time();

        }

        $this->SQLExec("INSERT INTO `WraithAPI_Sessions` (
                `assignedID`,
                `username`,
                `creatorIP`,
                `sessionToken`,
                `lastHeartbeatTime`
            ) VALUES (
                ?,
                ?,
                ?,
                ?,
                ?
            );",
            [
                $data["assignedID"],
                $data["username"],
                $data["creatorIP"],
                $data["sessionToken"],
                $data["lastHeartbeatTime"]
            ]
        );

        return $data["assignedID"];

    }

    // Delete a session
    function dbRemoveSessions($filter = [], $limit = -1, $offset = -1) {

        $validFilterColumnNames = [
            "assignedID",
            "username",
            "sessionToken",
            "lastHeartbeatTime"
        ];

        $SQL = "DELETE FROM `WraithAPI_Sessions`";

        $params = [];

        // Apply the filters
        $filterSQL = $this->generateFilter($filter, $validFilterColumnNames, $limit, $offset);
        $SQL .= $filterSQL[0];
        $params = array_merge($params, $filterSQL[1]);

        $statement = $this->SQLExec($SQL, $params);

    }

    // Get a list of all sessions
    function dbGetSessions($filter = [], $limit = -1, $offset = -1) {

        $validFilterColumnNames = [
            "assignedID",
            "username",
            "sessionToken",
            "lastHeartbeatTime"
        ];

        $SQL = "SELECT * FROM WraithAPI_Sessions";

        $params = [];

        // Apply the filters
        $filterSQL = $this->generateFilter($filter, $validFilterColumnNames, $limit, $offset);
        $SQL .= $filterSQL[0];
        $params = array_merge($params, $filterSQL[1]);

        $statement = $this->SQLExec($SQL, $params);

        // Get a list of sessions from the database
        $sessionsDB = $statement->fetchAll();

        $sessions = [];

        foreach ($sessionsDB as $session) {

            // Move the session ID to a separate variable
            $assignedID = $session["assignedID"];
            unset($session["assignedID"]);

            $sessions[$assignedID] = $session;

        }

        return $sessions;

    }

    // Update the session last heartbeat time
    function dbUpdateSessionLastHeartbeat($assignedID, $timeToSet = null) {

        // Set $timeToSet to the current time if no value was passed
        $timeToSet = isset($timeToSet) ? $timeToSet : time();

        // Update the last heartbeat time to the current time
        $SQL = "UPDATE WraithAPI_Sessions SET `lastHeartbeatTime` = ? WHERE `assignedID` = ?";

        $params = [
            $timeToSet,
            $assignedID
        ];

        $this->SQLExec($SQL, $params);

    }

    // Delete sessions which have not had a heartbeat recently
    function dbExpireSessions() {

        // The following cannot be easily done using dbRemoveSessions
        // as the current implementation of filters allows only for
        // exact matches and not greater-than / less-than. For this
        // reason, a separate SQL query is made. This can be replaced
        // if the filters implementation ever changes to allow comparisons
        // other than equals.

        // Remove all sessions where the last heartbeat time is older than
        // the $SETTINGS["managementSessionExpiryDelay"]
        $SQL = "DELETE FROM `WraithAPI_Sessions` WHERE `lastHeartbeatTime` < ?";

        $params = [
            // Unix timestamp for $SETTINGS["managementSessionExpiryDelay"] seconds ago
            time()-$this->dbGetSettings(["key" => ["managementSessionExpiryDelay"]])["managementSessionExpiryDelay"]
        ];

        $this->SQLExec($SQL, $params);

    }

    // EVENT TABLE MANAGEMENT (public)

    // Create/log an event
    function dbAddEvent($data) {

        // Check parameters and set defaults
        if (!(array_key_exists("assignedID", $data))) {

            $data["assignedID"] = uniqid();

        }
        if (!(array_key_exists("eventType", $data))) {

            // The eventType has no default value and is required
            return false;

        }
        if (!(array_key_exists("eventTargets", $data))) {

            // The eventTargets field is not required but has no defaults
            $data["eventTargets"] = [];

        }
        if (!(array_key_exists("eventTime", $data))) {

            $data["eventTime"] = time();

        }
        if (!(array_key_exists("eventData", $data))) {

            // eventData is required and has no default
            return false;

        }

        // Check that eventData is valid (has a description)
        if (!(hasKeys($data["eventData"], [
            "description"
        ]))) {

            return false;

        }

        $SQL = "INSERT INTO `WraithAPI_EventHistory` (
                `assignedID`,
                `eventType`,
                `eventTargets`,
                `eventTime`,
                `eventData`
            ) VALUES (
                ?,
                ?,
                ?,
                ?,
                ?
            )";

        $params = [
            $data["assignedID"],
            $data["eventType"],
            json_encode($data["eventTargets"]),
            $data["eventTime"],
            json_encode($data["eventData"])
        ];

        $this->SQLExec($SQL, $params);

        return $data["assignedID"];

    }

    function dbRemoveEvents($filter = [], $limit = -1, $offset = -1) {

        // TODO
        $validFilterColumnNames = [
            "assignedID",
            "hostProperties",
            "wraithProperties",
            "lastHeartbeatTime",
            "issuedCommands"
        ];

        $SQL = "DELETE FROM `WraithAPI_ActiveWraiths`";

        $params = [];

        // Apply the filters
        $filterSQL = $this->generateFilter($filter, $validFilterColumnNames, $limit, $offset);
        $SQL .= $filterSQL[0];
        $params = array_merge($params, $filterSQL[1]);

        $statement = $this->SQLExec($SQL, $params);

    }

    function dbGetEvents($filter = [], $limit = -1, $offset = -1) {

        // TODO
        $validFilterColumnNames = [
            "assignedID",
            "hostProperties",
            "wraithProperties",
            "lastHeartbeatTime",
            "issuedCommands"
        ];

        $SQL = "SELECT * FROM WraithAPI_ActiveWraiths";

        $params = [];

        // Apply the filters
        $filterSQL = $this->generateFilter($filter, $validFilterColumnNames, $limit, $offset);
        $SQL .= $filterSQL[0];
        $params = array_merge($params, $filterSQL[1]);

        $statement = $this->SQLExec($SQL, $params);

        // Get a list of wraiths from the database
        $wraithsDB = $statement->fetchAll();

        $wraiths = [];

        foreach ($wraithsDB as $wraith) {

            // Move the assigned ID to a separate variable
            $wraithID = $wraith["assignedID"];
            unset($wraith["assignedID"]);

            $wraiths[$wraithID] = $wraith;

        }

        return $wraiths;

    }

    // MISC (public)

    // Re-generate the switch encryption key for Wraiths
    function dbRegenWraithSwitchCryptKey($force = false) {

        if (!($force)) {

            // Separate if statements so the database is only read if needed

            // If there are active sessions
            $allWraiths = $this->dbGetWraiths();
            if (!(sizeof($allWraiths) === 0)) {

                return false;

            }

        }

        // Update the switch encryption key
        $this->dbSetSetting("wraithSwitchCryptKey", bin2hex(random_bytes(25)));

        return true;

    }

    // Re-generate the first-layer encryption key for management sessions
    function dbRegenMgmtCryptKey($force = false) {

        if (!($force)) {

            // Separate if statements so the database is only read if needed

            // If there are active sessions
            $allSessions = $this->dbGetSessions();
            if (!(sizeof($allSessions) === 0)) {

                return false;

            }

        }

        // Update the first layer encryption key
        $this->dbSetSetting("managementFirstLayerEncryptionKey", bin2hex(random_bytes(25)));

        return true;

    }

}
