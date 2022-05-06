/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import java.io.IOException;

public interface EventListener<T> {
    T next() throws InterruptedException, IOException;
    void close();
}
