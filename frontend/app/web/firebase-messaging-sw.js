importScripts("https://www.gstatic.com/firebasejs/10.7.0/firebase-app-compat.js");
importScripts("https://www.gstatic.com/firebasejs/10.7.0/firebase-messaging-compat.js");

firebase.initializeApp({
  apiKey: "AIzaSyAS0eqroCDyj7E87tUCVnUgwhXTNV9zv-g",
  authDomain: "ledgerguard-c7557.firebaseapp.com",
  projectId: "ledgerguard-c7557",
  storageBucket: "ledgerguard-c7557.firebasestorage.app",
  messagingSenderId: "761541341095",
  appId: "1:761541341095:web:ae2607c9547a090011d082",
});

const messaging = firebase.messaging();

// Handle background messages
messaging.onBackgroundMessage((message) => {
  console.log("[FCM SW] Background message:", message);
});
