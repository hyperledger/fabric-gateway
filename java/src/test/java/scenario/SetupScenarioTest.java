/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import org.junit.jupiter.api.Test;

public class SetupScenarioTest {
	@Test
	public void startFabric() throws Exception {
		System.err.println("Starting Fabric");
    	ScenarioSteps.startFabric();
	}
}
