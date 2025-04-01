// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";
import "../src/protofire-game.sol";

contract ProtofireGameTest is Test {
    ProtofireGame public game;

    function setUp() public {
        game = new ProtofireGame();
    }

    function testStoreGameResult() public {
        string memory player1 = "Ulad";
        string memory player2 = "Arsenii";
        uint8 winner = 1; // player1 wins

        game.storeGameResult(
            _stringToBytes15(player1),
            _stringToBytes15(player2),
            winner
        );

        (
            bytes15 resultPlayer1,
            bytes15 resultPlayer2,
            uint8 resultWinner
        ) = game.getGameResult(0);

        assertEq(
            resultPlayer1,
            _stringToBytes15(player1),
            "Player1 name mismatch"
        );
        assertEq(
            resultPlayer2,
            _stringToBytes15(player2),
            "Player2 name mismatch"
        );
        assertEq(resultWinner, winner, "Winner value mismatch");
    }

    function testGetTotalGames() public {
        assertEq(game.getTotalGames(), 0, "Initial game count should be zero");

        game.storeGameResult(
            _stringToBytes15("Ulad"),
            _stringToBytes15("Arsenii"),
            1 // Ulad (player1) wins
        );
        game.storeGameResult(
            _stringToBytes15("Arsenii"),
            _stringToBytes15("Ulad"),
            2 // Ulad (player2) wins
        );
        game.storeGameResult(
            _stringToBytes15("Ulad"),
            _stringToBytes15("Arsenii"),
            1 // Ulad (player1) wins
        );

        assertEq(game.getTotalGames(), 3, "Game count should be 3");
    }

    function testGetGameResultRevert() public {
        vm.expectRevert("Index out of bounds");
        game.getGameResult(0);
    }

    function testGameResultEvent() public {
        string memory player1 = "Alice";
        string memory player2 = "Bob";
        uint8 winner = 1; // Alice (player1) wins

        vm.expectEmit(true, true, true, true);
        emit ProtofireGame.GameResultStored(
            _stringToBytes15(player1),
            _stringToBytes15(player2),
            winner
        );

        game.storeGameResult(
            _stringToBytes15(player1),
            _stringToBytes15(player2),
            winner
        );
    }

    function _stringToBytes15(
        string memory source
    ) internal pure returns (bytes15 result) {
        bytes memory tempEmptyStringTest = bytes(source);
        if (tempEmptyStringTest.length == 0) {
            return 0x0;
        }

        assembly {
            result := mload(add(source, 32))
        }
    }
}
