// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract ProtofireGame {
    enum Move {
        Rock,
        Paper,
        Scissors
    }

    // 15 bytes is enough for a player name of 15 characters.
    // so the 2 players and the winner can fit in a single storage slot.
    struct GameResult {
        bytes15 player1;
        bytes15 player2;
        uint8 winner;
    }

    GameResult[] private gameResults;

    event GameResultStored(
        bytes15 indexed player1,
        bytes15 indexed player2,
        uint8 winner
    );

    function storeGameResult(
        bytes15 player1,
        bytes15 player2,
        uint8 winner
    ) external {
        gameResults.push(GameResult(player1, player2, winner));
        emit GameResultStored(player1, player2, winner);
    }

    function getTotalGames() external view returns (uint256) {
        return gameResults.length;
    }

    function getGameResult(
        uint256 index
    ) external view returns (bytes15, bytes15, uint8) {
        require(index < gameResults.length, "Index out of bounds");
        GameResult storage result = gameResults[index];
        return (result.player1, result.player2, result.winner);
    }
}
