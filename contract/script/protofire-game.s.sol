// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Script.sol";
import "../src/protofire-game.sol";

contract ProtofireGameScript is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");

        vm.startBroadcast(deployerPrivateKey);

        ProtofireGame game = new ProtofireGame();

        vm.stopBroadcast();
    }
}
