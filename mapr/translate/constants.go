/*
 * Copyright 2020-present Open Networking Foundation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package translate

// TODO: use P4 enums to generate values from p4info
const (
	IfTypeUnknown byte = 0x00
	IfTypeCore    byte = 0x01
	IfTypeAccess  byte = 0x02
)
