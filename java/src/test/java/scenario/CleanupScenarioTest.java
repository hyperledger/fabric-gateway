/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import org.junit.jupiter.api.Test;

public class CleanupScenarioTest {
	@Test
	public void stopFabric() throws Exception {
		ScenarioSteps.stopFabric();
	}
}
