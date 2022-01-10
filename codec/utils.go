// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "fmt"
    "io"

    "blockwatch.cc/tzgo/tezos"
)

// ensureTagAndSize reads the binary operation's tag and matches it against the expected
// type tag and minimum size for the operation under the current protocol. It returns
// an error when tag does not match or when the buffer is too short for reading the
// mandatory operation contents.
func ensureTagAndSize(buf *bytes.Buffer, typ tezos.OpType, ver int) error {
    if buf == nil {
        return io.ErrShortBuffer
    }

    tag, err := buf.ReadByte()
    if err != nil {
        // unread so the caller is able to repair
        buf.UnreadByte()
        return err
    }

    if tag != typ.TagVersion(ver) {
        // unread so the caller is able to repair
        buf.UnreadByte()
        return fmt.Errorf("invalid tag %d for op type %s", tag, typ)
    }

    // don't fail size checks for undefined ops
    sz := typ.MinSizeVersion(ver)
    if buf.Len() < sz-1 {
        fmt.Printf("short buffer for tag %d for op type %s: exp=%d got=%d\n", tag, typ,
            sz-1, buf.Len())
        buf.UnreadByte()
        return io.ErrShortBuffer
    }

    return nil
}

func readInt64(buf []byte) (int64, error) {
    if len(buf) != 8 {
        return 0, io.ErrShortBuffer
    }
    return int64(enc.Uint64(buf)), nil
}

func readInt32(buf []byte) (int32, error) {
    if len(buf) != 4 {
        return 0, io.ErrShortBuffer
    }
    return int32(enc.Uint32(buf)), nil
}

func readInt16(buf []byte) (int16, error) {
    if len(buf) != 2 {
        return 0, io.ErrShortBuffer
    }
    return int16(enc.Uint16(buf)), nil
}

func readUint32(buf []byte) (uint32, error) {
    if len(buf) != 4 {
        return 0, io.ErrShortBuffer
    }
    return enc.Uint32(buf), nil
}

func readUint16(buf []byte) (uint16, error) {
    if len(buf) != 2 {
        return 0, io.ErrShortBuffer
    }
    return enc.Uint16(buf), nil
}

func readBool(buf []byte) (bool, error) {
    if len(buf) != 1 {
        return false, io.ErrShortBuffer
    }
    return buf[0] == 255, nil
}

func readByte(buf []byte) (byte, error) {
    if len(buf) != 1 {
        return 0, io.ErrShortBuffer
    }
    return buf[0], nil
}
