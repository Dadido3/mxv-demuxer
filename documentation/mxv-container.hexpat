#pragma author David Vogel (Dadido3, D3)
#pragma description Magix Video File (MXV)
#pragma magic [4D 58 52 49 46 46 36 34] @ 0x00

bool DecodeAll in;      // When false: Will limit decoding of lists by "ListByteLimit" bytes, and will not decode pointers.
bool DecodePointers in; // When true: Will decode offsets in lookup tables as pointers. It's slow as it will decode chunks several times.
u64 ListByteLimit = 1000000;

u64 VideoFrames out;        // Number of detected video frames. There may be fewer frames shown here, as the format is able to deduplicate frames.
u64 AudioFrames out;        // Number of detected audio frames. There may be fewer frames shown here, as the format is able to deduplicate frames.
u64 AudioSamples out;       // Number of detected audio samples.
u64 VFTEEntries out;        // Number of detected VFTE entries. This corresponds to the real number of video frames.
u64 AFTEEntries out;        // Number of detected AFTE entries. This corresponds to the real number of audio frames.
u64 MXJVFT64Entries out;    // Number of detected VFT entries.
u64 MaxVideoChunkSize out;  // Size of the largest video frame chunk.
u64 MaxAudioChunkSize out;  // Size of the largest video frame chunk.

import hex.core;
import std.io;
import std.sys;

// Form type, aka "File type".
enum MXFormType : u64 {
    MXJVID64 = 0x34364449564A584D, // Magix Video File (MXV).
};

// Chunk32 identifiers.
enum MXIdentifier32 : u32 {
    MXVFTE = 0x45544656, // References to a video frame chunk (Video frame table entry?).
    MXAFTE = 0x45544641, // References to a audio frame chunk (Audio frame table entry?).
};

// Chunk64 identifiers.
enum MXIdentifier64 : u64 {
    MXRIFF64 = 0x343646464952584D, // Root chunk.
    MXLIST32 = 0x32335453494C584D, // List of sub-chunks.
    MXLIST64 = 0x34365453494C584D, // List of sub-chunks.

    MXJVAF64 = 0x34364641564A584D, // Audio frame data as raw PCM data.
    MXJVCO64 = 0x34364F43564A584D, // Guess: Video color something?
    MXJVFT64 = 0x34365446564A584D, // Video frame table: File offsets for fast seeking of specific frames. Contains one more entry than there are video frames in my test files. This is probably due to the list not containing any size information, and therefore the difference between the current and next entry is used. Therefore the entry for the last frame needs another entry after it.
    MXJVH264 = 0x34363248564A584D, // Video format info.
    MXJVHD64 = 0x34364448564A584D, // Older (and shorter) version of the video format info chunk.
    MXJVPD64 = 0x34364450564A584D,
    MXJVVF64 = 0x34364656564A584D, // Video frame data as a JPEG.
    MXWFMT64 = 0x3436544D4657584D, // Audio format info (Wave format).
};

// List content types.
enum MXContentType64 : u64 {
    MXJVFL64 = 0x34364C46564A584D, // List of video and audio frames. ("Video Frame List"?)
    MXJVTL32 = 0x32334C54564A584D, // List of offsets of the respective video and audio chunks.
};

using Chunk64;

struct ChunkMXAFTE {
    if (DecodePointers) {Chunk64 *AudioFrameChunk: u64;} else {u64 AudioFrameChunkOffset;}
    u32 ChunkSize;
    u64 StartSample;
    u32 Samples;

    AFTEEntries += 1;
};

struct ChunkMXVFTE {
    if (DecodePointers) {Chunk64 *VideoFrameChunk: u64;} else {u64 VideoFrameChunkOffset;}
    u32 ChunkSize;

    VFTEEntries += 1;
};

struct Chunk32 {
    MXIdentifier32 Identifier;
    u32 Length;

    u64 StartPos = $;

    match (Identifier) {
        (MXIdentifier32::MXAFTE): ChunkMXAFTE [[inline]];
        (MXIdentifier32::MXVFTE): ChunkMXVFTE [[inline]];
        (_): {
            std::print("Unknown MXIdentifier32 0x{:X}.", u64(Identifier));
            $ += Length; // Jump over unknown chunks.
        }
    }

    u64 ActualLength = $ - StartPos;
    std::assert(ActualLength == Length, std::format("Chunk {} has a total length of {}, expected {}.", Identifier, ActualLength, Length));
};

struct ChunkMXRIFF64 {
    Chunk64 Chunks[while($ - parent.StartPos < parent.Length)];
};

struct ChunkMXLIST32 {
    Chunk32 Chunks[while($ - parent.StartPos < parent.Length && ($ - parent.StartPos < ListByteLimit || DecodeAll))];
};

struct ChunkMXLIST64 {
    Chunk64 Chunks[while($ - parent.StartPos < parent.Length && ($ - parent.StartPos < ListByteLimit || DecodeAll))];
};

struct ChunkMXJVAF64 {
    u32 ChannelBitDepth;
    u64 StartSample;
    u32 Samples;
    u8 PCMData[parent.Length - 16];

    AudioFrames += 1;
    AudioSamples += Samples;
    if (MaxAudioChunkSize < parent.Length) {MaxAudioChunkSize = parent.Length;}
};

struct ChunkMXJVCO64 {
    u8 Data[24];
};

struct ChunkMXJVFT64 {
    u64 VideoFrameOffset[parent.Length/8];

    MXJVFT64Entries += parent.Length/8;
};

struct ChunkMXJVH264 {
    u32;                    // Seems to be always 112, may be the size of this struct.
    u32;                    // Guess: May be some sort of version. (File or encoding software)
    if (DecodePointers) {Chunk64 *FrameTable: u64;} else {u64 FrameTableOffset;}  // File offset of ChunkMXJVFT64.
    u64 VideoFrames;        // The total number of VFTE entries (These point to MXJVVF64 chunks that contain images). It's possible that several VFTE entries point to the same MXJVVF64 chunk.
    u32 MaxReadSize;        // Guess: Seems to be the max of all video and audio frame chunk pairs. Perhaps to tell any software what the max. buffer size needs to be.
    u32;                    // Guess: May be some sort of version. (File or encoding software)
    u64;
    double Framerate;       // Number of frames per second. For interlaced video it will store the number of full frames per second. I.e. PAL has 50 fields/s and therefore 25 full frames/s.
    u32 FrameWidth;
    u32 FrameHeight;
    u32 FrameWidth2;        // Maybe needed when the video is anamorphic? It's the same as the above width in all my test files.
    u32 FrameHeight2;       // Maybe needed when the video is anamorphic? It's the same as the above height in all my test files.
    u32;
    u32 MaxJPEGSize;        // The JPEG data size of the largest MXJVVF64 chunk. (The size only includes the JPEG data)
    u64 AudioFrames;        // Number of AFTE entries (These point to MXJVAF64 chunks that contain waveform data)
    u64 MaxAudioChunkSize;  // The size of the largest MXJVAF64 chunk. (The size includes the chunk identifier and length field and all its data)
    double AspectRatio;     // Final image aspect ratio. If this ratio != FrameWidth / FrameHeight the video doesn't have square pixels.
    u64;                    // Guess: This could be audio related stuff.
    u64 AudioSamples;       // Total number of audio samples.
};

struct ChunkMXJVHD64 {
    u32;
    u32;
    if (DecodePointers) {Chunk64 *FrameTable: u64;} else {u64 FrameTableOffset;}
    u64 VideoFrames;
    u32 SeekTableMaxReadSize;
    u32;
    u64;
    double Framerate;
    u32 FrameWidth;
    u32 FrameHeight;
    u32 FrameWidth2;
    u32 FrameHeight2;
    u32;
    u32 MaxJPEGSize;
};

struct ChunkMXJVPD64 {
    u8 Data[20];
};

struct ChunkMXJVVF64 {
    u8 JPEGData[while($ - parent.StartPos < parent.Length)];

    VideoFrames += 1;
    if (MaxVideoChunkSize < parent.Length) {MaxVideoChunkSize = parent.Length;}
};

struct ChunkMXWFMT64 {
    u16 Tracks;             // Guess: The number of tracks if there are multiple. But it could also encode the byte format used.
    u16 Channels;           // The number of channels.
    u32 SampleRate;         // Samples per second.
    u32 ByteRate;           // Bytes per second.
    u16 BytesPerSample;     // The number of bytes per sample. (Channels * ChannelBitDepth / 8) or (ByteRate / SampleRate)
    u32 ChannelBitDepth;    // Bits per channel per sample.
};

struct Chunk64 {
    MXIdentifier64 Identifier;
    u64 Length;
    
    match (Identifier) {
        (MXIdentifier64::MXRIFF64): MXFormType FormType;
        (MXIdentifier64::MXLIST32 | MXIdentifier64::MXLIST64): MXContentType64 ContentType;
    }

    u64 StartPos = $;

    match (Identifier) {
        (MXIdentifier64::MXRIFF64): {ChunkMXRIFF64 [[inline]]; $ = StartPos + Length;}
        (MXIdentifier64::MXLIST32): {ChunkMXLIST32 [[inline]]; $ = StartPos + Length;}
        (MXIdentifier64::MXLIST64): {ChunkMXLIST64 [[inline]]; $ = StartPos + Length;}
        (MXIdentifier64::MXJVAF64): ChunkMXJVAF64 [[inline]];
        (MXIdentifier64::MXJVCO64): ChunkMXJVCO64 [[inline]];
        (MXIdentifier64::MXJVFT64): ChunkMXJVFT64 [[inline]];
        (MXIdentifier64::MXJVH264): ChunkMXJVH264 [[inline]];
        (MXIdentifier64::MXJVHD64): ChunkMXJVHD64 [[inline]];
        (MXIdentifier64::MXJVPD64): ChunkMXJVPD64 [[inline]];
        (MXIdentifier64::MXJVVF64): ChunkMXJVVF64 [[inline]];
        (MXIdentifier64::MXWFMT64): ChunkMXWFMT64 [[inline]];
        (_): {
            std::print("Unknown MXIdentifier64 0x{:X}.", u64(Identifier));
            $ += Length; // Jump over unknown chunks.
        }
    }

    u64 ActualLength = $ - StartPos;
    std::assert(ActualLength == Length, std::format("Chunk {} has a total length of {}, expected {}.", Identifier, ActualLength, Length));
};

Chunk64 RootChunk @ 0x00;