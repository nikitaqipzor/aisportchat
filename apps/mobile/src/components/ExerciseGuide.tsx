import React from 'react';
import {Modal, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {Exercise} from '../api/client';

export function ExerciseGuide({exercise, visible, onClose}: {exercise: Exercise | null; visible: boolean; onClose: () => void}) {
  if (!exercise) return null;
  return (
    <Modal visible={visible} animationType="slide" presentationStyle="pageSheet" onRequestClose={onClose}>
      <ScrollView contentContainerStyle={styles.container}>
        <View style={styles.header}>
          <View style={{flex: 1}}>
            <Text style={styles.eyebrow}>ТЕХНИКА</Text>
            <Text style={styles.title}>{exercise.name}</Text>
          </View>
          <Pressable onPress={onClose} style={styles.close}><Text style={styles.closeText}>✕</Text></Pressable>
        </View>
        <Text style={styles.description}>{exercise.description}</Text>
        <Section title="Как выполнять" items={exercise.instructions} numbered />
        <Section title="Частые ошибки" items={exercise.common_mistakes} />
        <Section title="Подсказки" items={exercise.technique_tips} />
        <View style={styles.note}><Text style={styles.noteText}>Если движение вызывает резкую или необычную боль — прекрати подход и не пытайся «продавить» амплитуду.</Text></View>
      </ScrollView>
    </Modal>
  );
}

function Section({title, items, numbered = false}: {title: string; items: string[]; numbered?: boolean}) {
  return <View style={styles.section}>
    <Text style={styles.sectionTitle}>{title}</Text>
    {items.map((item, index) => <Text key={`${title}-${index}`} style={styles.item}>{numbered ? `${index + 1}.` : '•'} {item}</Text>)}
  </View>;
}

const styles = StyleSheet.create({
  container: {padding: 22, gap: 18},
  header: {flexDirection: 'row', alignItems: 'flex-start', gap: 12},
  eyebrow: {fontSize: 12, fontWeight: '800', letterSpacing: 1.4, opacity: 0.5},
  title: {fontSize: 30, lineHeight: 35, fontWeight: '900', marginTop: 4},
  close: {width: 42, height: 42, borderWidth: 1, borderRadius: 21, alignItems: 'center', justifyContent: 'center'},
  closeText: {fontSize: 18, fontWeight: '800'},
  description: {fontSize: 16, lineHeight: 24, opacity: 0.7},
  section: {gap: 9, borderTopWidth: 1, paddingTop: 16},
  sectionTitle: {fontSize: 19, fontWeight: '800'},
  item: {fontSize: 15, lineHeight: 22},
  note: {borderWidth: 1, borderRadius: 16, padding: 14},
  noteText: {fontSize: 14, lineHeight: 20, fontWeight: '600'},
});
