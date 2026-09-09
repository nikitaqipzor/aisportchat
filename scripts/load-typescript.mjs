import fs from 'node:fs';
import path from 'node:path';
import {execFileSync} from 'node:child_process';
import {pathToFileURL} from 'node:url';

export async function loadTypeScript() {
  const projectEntry = path.resolve('apps/mobile/node_modules/typescript/lib/typescript.js');
  if (fs.existsSync(projectEntry)) return import(pathToFileURL(projectEntry).href);
  try { return await import('typescript'); } catch { /* try active global npm below */ }
  const globalRoot = execFileSync('npm', ['root', '-g'], {encoding: 'utf8'}).trim();
  return import(pathToFileURL(path.join(globalRoot, 'typescript/lib/typescript.js')).href);
}
