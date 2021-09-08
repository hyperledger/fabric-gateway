/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.UncheckedIOException;
import java.security.GeneralSecurityException;
import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.StreamSupport;

import org.bouncycastle.asn1.ASN1Encodable;
import org.bouncycastle.asn1.ASN1InputStream;
import org.bouncycastle.asn1.ASN1Integer;
import org.bouncycastle.asn1.ASN1Primitive;
import org.bouncycastle.asn1.ASN1Sequence;
import org.bouncycastle.asn1.DERSequenceGenerator;

final class ECSignature {
    private final ASN1Integer r;
    private final ASN1Integer s;

    static ECSignature fromBytes(final byte[] derSignature) throws GeneralSecurityException {
        try (ByteArrayInputStream inStream = new ByteArrayInputStream(derSignature);
             ASN1InputStream asnInputStream = new ASN1InputStream(inStream)) {
            ASN1Primitive asn1 = asnInputStream.readObject();

            if (!(asn1 instanceof ASN1Sequence)) {
                throw new GeneralSecurityException("Invalid signature type: " + asn1.getClass().getTypeName());
            }

            ASN1Sequence asn1Sequence = (ASN1Sequence) asn1;
            List<ASN1Integer> signatureParts = StreamSupport.stream(asn1Sequence.spliterator(), false)
                    .map(ASN1Encodable::toASN1Primitive)
                    .filter(asn1Primitive -> asn1Primitive instanceof ASN1Integer)
                    .map(asn1Primitive -> (ASN1Integer) asn1Primitive)
                    .collect(Collectors.toList());
            if (signatureParts.size() != 2) {
                throw new GeneralSecurityException("Invalid signature. Expected 2 values but got " + signatureParts.size());
            }

            return new ECSignature(signatureParts.get(0), signatureParts.get(1));
        } catch (IOException e) {
            // Should not happen reading from ByteArrayInputStream
            throw new UncheckedIOException(e);
        }
    }

    ECSignature(final ASN1Integer r, final ASN1Integer s) {
        this.r = r;
        this.s = s;
    }

    public ASN1Integer getR() {
        return r;
    }

    public ASN1Integer getS() {
        return s;
    }

    public byte[] getBytes() {
        try (ByteArrayOutputStream bytesOut = new ByteArrayOutputStream()) {
            DERSequenceGenerator sequence = new DERSequenceGenerator(bytesOut);
            sequence.addObject(r);
            sequence.addObject(s);
            sequence.close();
            return bytesOut.toByteArray();
        } catch (IOException e) {
            // Should not happen writing to ByteArrayOutputStream
            throw new UncheckedIOException(e);
        }
    }
}
