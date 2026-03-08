import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../core/theme/app_theme.dart';
import '../../domain/entities/chat_message.dart';
import '../blocs/chat/chat.dart';
import '../widgets/chat/data_panel.dart';
import '../widgets/chat/message_bubble.dart';

class ChatPage extends StatefulWidget {
  const ChatPage({super.key});

  @override
  State<ChatPage> createState() => _ChatPageState();
}

class _ChatPageState extends State<ChatPage> {
  final _controller = TextEditingController();
  final _scrollController = ScrollController();
  final _focusNode = FocusNode();

  @override
  void dispose() {
    _controller.dispose();
    _scrollController.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _sendMessage() {
    final text = _controller.text.trim();
    if (text.isEmpty) return;

    context.read<ChatBloc>().add(SendMessageRequested(text));
    _controller.clear();
    _focusNode.requestFocus();
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('AI Assistant'),
        actions: [
          IconButton(
            icon: const Icon(Icons.delete_outline),
            tooltip: 'Clear chat',
            onPressed: () {
              context.read<ChatBloc>().add(const ClearChatRequested());
            },
          ),
        ],
      ),
      body: BlocConsumer<ChatBloc, ChatState>(
        listener: (context, state) {
          if (state is ChatLoaded) {
            _scrollToBottom();
          }
        },
        builder: (context, state) {
          if (state is ChatInitial) {
            return _buildWelcome(context);
          }

          final messages = state is ChatLoaded
              ? state.messages
              : state is ChatError
                  ? state.previousMessages
                  : <ChatMessage>[];

          final isStreaming =
              state is ChatLoaded && state.isStreaming;

          final activeToolName =
              state is ChatLoaded ? state.activeToolName : null;

          // Find the latest assistant message with data state
          final dataMessage = _findLatestDataMessage(messages);

          return LayoutBuilder(
            builder: (context, constraints) {
              final isWide = constraints.maxWidth >= 800;

              if (isWide && dataMessage != null) {
                return Row(
                  children: [
                    Expanded(
                      flex: 3,
                      child: _buildChatPane(
                          context, messages, isStreaming, activeToolName, state),
                    ),
                    Expanded(
                      flex: 2,
                      child: DataPanel(message: dataMessage),
                    ),
                  ],
                );
              }

              return _buildChatPane(
                  context, messages, isStreaming, activeToolName, state);
            },
          );
        },
      ),
    );
  }

  Widget _buildWelcome(BuildContext context) {
    final suggestions = [
      'What is my MRR?',
      'Show at-risk stores',
      'Revenue breakdown this month',
    ];

    return Column(
      children: [
        Expanded(
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    Icons.smart_toy_outlined,
                    size: 48,
                    color: AppTheme.primary.withValues(alpha: 0.6),
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'Revenue Intelligence Assistant',
                    style: Theme.of(context).textTheme.titleLarge?.copyWith(
                          fontWeight: FontWeight.w600,
                        ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Ask about your subscriptions, revenue, risk metrics, and more.',
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: Theme.of(context).colorScheme.onSurfaceVariant,
                        ),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 24),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    alignment: WrapAlignment.center,
                    children: suggestions.map((s) {
                      return ActionChip(
                        label: Text(s),
                        onPressed: () {
                          context.read<ChatBloc>().add(SendMessageRequested(s));
                        },
                        backgroundColor:
                            AppTheme.primary.withValues(alpha: 0.08),
                        side: BorderSide(
                          color: AppTheme.primary.withValues(alpha: 0.2),
                        ),
                      );
                    }).toList(),
                  ),
                ],
              ),
            ),
          ),
        ),
        _buildInput(context, false),
      ],
    );
  }

  Widget _buildChatPane(
    BuildContext context,
    List<ChatMessage> messages,
    bool isStreaming,
    String? activeToolName,
    ChatState state,
  ) {
    return Column(
      children: [
        if (state is ChatError)
          MaterialBanner(
            content: Text(state.message),
            backgroundColor: AppTheme.danger.withValues(alpha: 0.1),
            leading: const Icon(Icons.error_outline, color: AppTheme.danger),
            actions: [
              TextButton(
                onPressed: () {
                  // Dismiss by transitioning back to loaded
                  if (state.previousMessages.isNotEmpty) {
                    context.read<ChatBloc>().add(
                          SendMessageRequested(
                              state.previousMessages.last.content),
                        );
                  }
                },
                child: const Text('Retry'),
              ),
            ],
          ),
        Expanded(
          child: ListView.builder(
            controller: _scrollController,
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: messages.length,
            itemBuilder: (context, index) {
              return MessageBubble(
                message: messages[index],
                onSuggestionTapped: (s) {
                  context.read<ChatBloc>().add(SuggestionTapped(s));
                },
              );
            },
          ),
        ),
        if (activeToolName != null)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
            child: Row(
              children: [
                SizedBox(
                  width: 12,
                  height: 12,
                  child: CircularProgressIndicator(
                    strokeWidth: 1.5,
                    color: AppTheme.primary,
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  'Querying ${_formatToolName(activeToolName)}...',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                      ),
                ),
              ],
            ),
          ),
        _buildInput(context, isStreaming),
      ],
    );
  }

  Widget _buildInput(BuildContext context, bool isStreaming) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 8, 12, 12),
      decoration: BoxDecoration(
        color: Theme.of(context).scaffoldBackgroundColor,
        border: Border(
          top: BorderSide(color: Theme.of(context).dividerColor),
        ),
      ),
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _controller,
              focusNode: _focusNode,
              enabled: !isStreaming,
              textInputAction: TextInputAction.send,
              onSubmitted: (_) => _sendMessage(),
              decoration: InputDecoration(
                hintText: isStreaming
                    ? 'Waiting for response...'
                    : 'Ask about your revenue data...',
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 10,
                ),
                isDense: true,
              ),
              maxLines: 3,
              minLines: 1,
            ),
          ),
          const SizedBox(width: 8),
          IconButton.filled(
            onPressed: isStreaming ? null : _sendMessage,
            icon: const Icon(Icons.send, size: 20),
            style: IconButton.styleFrom(
              backgroundColor: AppTheme.primary,
              disabledBackgroundColor:
                  AppTheme.primary.withValues(alpha: 0.3),
            ),
          ),
        ],
      ),
    );
  }

  ChatMessage? _findLatestDataMessage(List<ChatMessage> messages) {
    for (int i = messages.length - 1; i >= 0; i--) {
      if (messages[i].isAssistant && messages[i].hasDataState) {
        return messages[i];
      }
    }
    return null;
  }

  String _formatToolName(String toolName) {
    // "risk__get_risk_summary" → "risk summary"
    final parts = toolName.split('__');
    final name = parts.length > 1 ? parts[1] : parts[0];
    return name.replaceAll('_', ' ').replaceAll('get ', '');
  }
}
