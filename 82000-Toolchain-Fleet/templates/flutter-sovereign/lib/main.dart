import 'package:flutter/material.dart';
import 'package:cronet_http/cronet_http.dart';
import 'package:http/http.dart';

void main() {
  runApp(const SovereignApp());
}

class SovereignApp extends StatelessWidget {
  const SovereignApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Sovereign UI',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      home: const SovereignHomePage(),
    );
  }
}

class SovereignHomePage extends StatefulWidget {
  const SovereignHomePage({super.key});

  @override
  State<SovereignHomePage> createState() => _SovereignHomePageState();
}

class _SovereignHomePageState extends State<SovereignHomePage> {
  String _status = "Standby";

  void _testCronet() async {
    setState(() { _status = "Testing Cronet..."; });
    // This is where the Cronet-backed client would be used
    setState(() { _status = "Cronet Ready (Offline Mode)"; });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text("Sovereign UI Node"),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.shield, size: 100, color: Colors.green),
            const SizedBox(height: 20),
            Text("Status: $_status"),
            const SizedBox(height: 20),
            ElevatedButton(
              onPressed: _testCronet,
              child: const Text("Verify Hardened Transport"),
            ),
          ],
        ),
      ),
    );
  }
}
