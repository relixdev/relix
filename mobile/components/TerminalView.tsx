import React, { useRef, useEffect } from 'react';
import { ScrollView, Text, View, StyleSheet, Platform } from 'react-native';

const MONO_FONT = Platform.select({
  ios: 'Courier New',
  android: 'monospace',
  default: 'monospace',
}) as string;

// ANSI foreground color codes → hex colors
const ANSI_FG: Record<number, string> = {
  30: '#000000',
  31: '#FF3B30',
  32: '#34C759',
  33: '#FF9500',
  34: '#007AFF',
  35: '#AF52DE',
  36: '#5AC8FA',
  37: '#EBEBF0',
  90: '#636366',
  91: '#FF6961',
  92: '#30D158',
  93: '#FFD60A',
  94: '#409CFF',
  95: '#DA8FFF',
  96: '#70D7FF',
  97: '#FFFFFF',
};

interface AnsiSegment {
  text: string;
  color?: string;
  bold?: boolean;
}

function parseAnsi(line: string): AnsiSegment[] {
  const segments: AnsiSegment[] = [];
  const regex = /\x1b\[([0-9;]*)m/g;
  let lastIndex = 0;
  let currentColor: string | undefined;
  let bold = false;

  let match: RegExpExecArray | null;
  while ((match = regex.exec(line)) !== null) {
    const before = line.slice(lastIndex, match.index);
    if (before) {
      segments.push({ text: before, color: currentColor, bold });
    }
    lastIndex = match.index + match[0].length;

    const codes = match[1].split(';').map(Number);
    for (const code of codes) {
      if (code === 0) {
        currentColor = undefined;
        bold = false;
      } else if (code === 1) {
        bold = true;
      } else if (code === 22) {
        bold = false;
      } else if (ANSI_FG[code] !== undefined) {
        currentColor = ANSI_FG[code];
      }
    }
  }

  const remaining = line.slice(lastIndex);
  if (remaining) {
    segments.push({ text: remaining, color: currentColor, bold });
  }

  if (segments.length === 0) {
    segments.push({ text: line });
  }

  return segments;
}

interface TerminalViewProps {
  lines: string[];
}

export default function TerminalView({ lines }: TerminalViewProps) {
  const scrollRef = useRef<ScrollView>(null);

  useEffect(() => {
    scrollRef.current?.scrollToEnd({ animated: true });
  }, [lines]);

  return (
    <ScrollView
      ref={scrollRef}
      style={styles.container}
      contentContainerStyle={styles.content}
      onContentSizeChange={() => scrollRef.current?.scrollToEnd({ animated: false })}
    >
      {lines.map((line, i) => {
        const segments = parseAnsi(line);
        return (
          <View key={i} style={styles.lineRow}>
            {segments.map((seg, j) => (
              <Text
                key={j}
                style={[
                  styles.lineText,
                  seg.color != null ? { color: seg.color } : null,
                  seg.bold ? styles.bold : null,
                ]}
              >
                {seg.text}
              </Text>
            ))}
          </View>
        );
      })}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1a1a2e',
  },
  content: {
    padding: 12,
    paddingBottom: 24,
  },
  lineRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginBottom: 1,
  },
  lineText: {
    fontFamily: MONO_FONT,
    fontSize: 13,
    lineHeight: 19,
    color: '#E0E0E0',
  },
  bold: {
    fontWeight: '700',
  },
});
