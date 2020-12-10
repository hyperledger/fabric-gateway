/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

module.exports.connect = require('./gateway').connect;
module.exports.Gateway = require('./gateway').Gateway;
module.exports.Signers = require('./identity/signers');
