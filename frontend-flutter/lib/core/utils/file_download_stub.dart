import 'dart:typed_data';

/// Non-web fallback. There is no ambient browser to hand the file to, so the
/// caller is expected to surface a graceful message. Returns false so callers
/// can show a "not supported on this platform" SnackBar.
bool downloadBytes(Uint8List bytes, String filename, String mimeType) {
  return false;
}
