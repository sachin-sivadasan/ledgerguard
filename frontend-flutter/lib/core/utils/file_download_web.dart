// This helper is only ever loaded on Flutter web via the conditional import in
// file_download.dart, so the web-only library warnings are expected here.
// ignore_for_file: avoid_web_libraries_in_flutter, deprecated_member_use
import 'dart:typed_data';
import 'dart:html' as html;

/// Triggers a client-side download in the browser using a Blob + object URL.
bool downloadBytes(Uint8List bytes, String filename, String mimeType) {
  final blob = html.Blob(<Object>[bytes], mimeType);
  final url = html.Url.createObjectUrlFromBlob(blob);
  try {
    html.AnchorElement(href: url)
      ..setAttribute('download', filename)
      ..click();
  } finally {
    html.Url.revokeObjectUrl(url);
  }
  return true;
}
