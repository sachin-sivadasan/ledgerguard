import 'dart:typed_data';

import 'file_download_stub.dart'
    if (dart.library.html) 'file_download_web.dart' as impl;

/// Triggers a client-side download of [bytes] as [filename].
///
/// On web this creates a Blob + object URL and clicks a synthetic anchor.
/// On non-web platforms this is a no-op and returns false so the caller can
/// show a graceful message.
bool downloadBytes(Uint8List bytes, String filename, String mimeType) =>
    impl.downloadBytes(bytes, filename, mimeType);
