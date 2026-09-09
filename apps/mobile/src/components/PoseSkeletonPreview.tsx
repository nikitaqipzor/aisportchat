import React from 'react';
import {StyleSheet, Text, View} from 'react-native';
import {PoseFrame} from '../native/techniqueVideo';

const W = 230, H = 300;
const LINKS: Array<[number, number]> = [[11,12],[11,13],[13,15],[12,14],[14,16],[11,23],[12,24],[23,24],[23,25],[25,27],[24,26],[26,28]];
const POINTS = [11,12,13,14,15,16,23,24,25,26,27,28];

export function PoseSkeletonPreview({frame}: {frame: PoseFrame}) {
  const l=frame.landmarks;
  return <View style={styles.wrap}><Text style={styles.label}>POSE LANDMARKS · кадр {Math.round(frame.timestamp_ms/100)/10} с</Text><View style={styles.canvas}>
    {LINKS.map(([a,b]) => <Bone key={`${a}-${b}`} a={l[a]} b={l[b]} />)}
    {POINTS.map(i => l[i] && <View key={i} style={[styles.point,{left:l[i].x*W-4,top:l[i].y*H-4,opacity:Math.max(.2,l[i].visibility)}]} />)}
  </View></View>;
}
function Bone({a,b}:{a?:{x:number;y:number;visibility:number};b?:{x:number;y:number;visibility:number}}) {
  if(!a||!b||a.visibility<.25||b.visibility<.25)return null;
  const x1=a.x*W,y1=a.y*H,x2=b.x*W,y2=b.y*H,dx=x2-x1,dy=y2-y1,len=Math.hypot(dx,dy),angle=Math.atan2(dy,dx),mx=(x1+x2)/2,my=(y1+y2)/2;
  return <View style={[styles.bone,{left:mx-len/2,top:my-1,width:len,opacity:Math.min(a.visibility,b.visibility),transform:[{rotate:`${angle}rad`}]}]} />;
}
const styles=StyleSheet.create({wrap:{alignItems:'center',gap:8},label:{fontSize:10,fontWeight:'800',letterSpacing:.7,color:'#777'},canvas:{width:W,height:H,backgroundColor:'#f5f5f5',borderRadius:22,overflow:'hidden'},bone:{position:'absolute',height:2,backgroundColor:'#111',borderRadius:2},point:{position:'absolute',width:8,height:8,borderRadius:4,backgroundColor:'#111'}});
