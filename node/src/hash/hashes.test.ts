/*
 * Copyright 2023 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as hashes from './hashes';

describe('hashes', () => {
    Object.entries(hashes).forEach(([name, hash]) => {
        describe(`${name}`, () => {
            it('Hashes of identical data are identical', () => {
                const message = Buffer.from('foobar');

                const hash1 = hash(message);
                const hash2 = hash(message);

                expect(hash1).toEqual(hash2);
            });

            it('Hashes of different data are not identical', () => {
                const foo = Buffer.from('foo');
                const bar = Buffer.from('bar');

                const fooHash = hash(foo);
                const barHash = hash(bar);

                expect(fooHash).not.toEqual(barHash);
            });
        });
    });

    it('none returns input', () => {
        const message = Buffer.from('foobar');
        const hash = hashes.none(message);
        expect(hash).toEqual(message);
    });
});
