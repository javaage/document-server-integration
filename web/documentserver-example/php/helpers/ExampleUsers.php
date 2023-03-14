<?php

namespace OnlineEditorsExamplePhp\Helpers;

use function OnlineEditorsExamplePhp\sendlog;

/**
 * (c) Copyright Ascensio System SIA 2023
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

final class ExampleUsers
{
    private array $descr_user_1;
    private array $descr_user_2;
    private array $descr_user_3;
    private array $descr_user_0;
    private array $users;

    public function __construct()
    {
        $this->descr_user_1 = [
            "File author by default",
            "Doesn’t belong to any group",
            "Can review all the changes",
            "Can perform all actions with comments",
            "The file favorite state is undefined",
            "Can create files from templates using data from the editor",
            "Can see the information about all users",
        ];
        $this->descr_user_2 = [
            "Belongs to Group2",
            "Can review only his own changes or changes made by users with no group",
            "Can view comments, edit his own comments and comments left by users with no group. 
        Can remove his own comments only",
            "This file is marked as favorite",
            "Can create new files from the editor",
            "Can see the information about users from Group2 and users who don’t belong to any group",
        ];
        $this->descr_user_3 = [
            "Belongs to Group3",
            "Can review changes made by Group2 users",
            "Can view comments left by Group2 and Group3 users. Can edit comments left by the Group2 users",
            "This file isn’t marked as favorite",
            "Can’t copy data from the file to clipboard",
            "Can’t download the file",
            "Can’t print the file",
            "Can create new files from the editor",
            "Can see the information about Group2 users",
        ];
        $this->descr_user_0 = [
            "The name is requested when the editor is opened",
            "Doesn’t belong to any group",
            "Can review all the changes",
            "Can perform all actions with comments",
            "The file favorite state is undefined",
            "Can't mention others in comments",
            "Can't create new files from the editor",
            "Can’t see anyone’s information",
            "Can't rename files from the editor",
            "Can't view chat",
            "View file without collaboration",
        ];
        $this->users = [
            new Users(
                "uid-1",
                "John Smith",
                "smith@example.com",
                "",
                null,
                [],
                null,
                null,
                [],
                $this->descr_user_1,
                true
            ),
            new Users(
                "uid-2",
                "Mark Pottato",
                "pottato@example.com",
                "group-2",
                ["group-2", ""],
                [
                    "view" => "",
                    "edit" => ["group-2", ""],
                    "remove" => ["group-2"],
                ],
                ["group-2", ""],
                true,
                [],
                $this->descr_user_2,
                false
            ),
            new Users(
                "uid-3",
                "Hamish Mitchell",
                "mitchell@example.com",
                "group-3",
                ["group-2"],
                [
                    "view" => ["group-3", "group-2"],
                    "edit" => ["group-2"],
                    "remove" => [],
                ],
                ["group-2"],
                false,
                ["copy", "download", "print"],
                $this->descr_user_3,
                false
            ),
            new Users(
                "uid-0",
                null,
                null,
                "",
                null,
                [],
                [],
                null,
                ["protect"],
                $this->descr_user_0,
                false
            ),
        ];
    }

    /**
     * Get a list of all the users
     *
     * @return array
     */
    public function getAllUsers(): array
    {
        return $this->users;
    }

    /**
     * Get a user by id specified
     *
     * @param string|null $id
     *
     * @return Users
     */
    public function getUser(?string $id): Users
    {
        foreach ($this->users as $user) {
            if ($user->id == $id) {
                sendlog("User ". $user->id, "common.log");
                return $user;
            }
        }
        return $this->users[0];
    }

    /**
     * Get a list of users with their id, name and email
     *
     * @param string|null $id
     *
     * @return array
     */
    public function getUserList(?string $id): array
    {
        $usersData = [];
        foreach ($this->users as $user) {
            if ($user->id != $id && $user->name != null && $user->email != null) {
                $usersData[] = [
                    "id" => $user->id,
                    "name" => $user->name,
                    "email" => $user->email,
                ];
            }
        }
        return $usersData;
    }
}
