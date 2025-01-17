import {Component, ElementRef, inject, input, OnDestroy, OnInit} from '@angular/core';
import {toObservable} from '@angular/core/rxjs-interop';
import {FormControl} from '@angular/forms';
import {defaultKeymap, history, historyKeymap, indentWithTab} from '@codemirror/commands';
import {yaml} from '@codemirror/lang-yaml';
import {HighlightStyle, indentOnInput, syntaxHighlighting} from '@codemirror/language';
import {Transaction} from '@codemirror/state';
import {EditorView, highlightSpecialChars, keymap} from '@codemirror/view';
import {tags} from '@lezer/highlight';
import {EMPTY, Subject, switchMap, takeUntil, tap} from 'rxjs';

@Component({
  selector: 'app-yaml-editor',
  template: '',
})
export class YamlEditorComponent implements OnInit, OnDestroy {
  // can not be named 'formControl' because of name conflict
  public readonly formCtrl = input<FormControl>();

  private readonly host = inject(ElementRef);
  private view!: EditorView;
  private readonly destroyed$ = new Subject<void>();

  constructor() {
    toObservable(this.formCtrl)
      .pipe(
        tap(x => console.log('1 formCtrl', x)),
        takeUntil(this.destroyed$),
        switchMap((fc) => fc?.valueChanges ?? EMPTY)
      )
      .subscribe((value) => {
        console.log('1.1 formCtrl valueChanges', value);
        this.view.state.update({changes: {from: 0, to: this.view.state.doc.length, insert: value ?? ''}});
        console.log('1.2 view updated')
      });
  }

  public ngOnInit(): void {
    this.view = new EditorView({
      doc: this.formCtrl()?.value ?? '',
      extensions: [
        indentOnInput(),
        history(),
        syntaxHighlighting(
          HighlightStyle.define([
            // TODO: Improve highlight style
            {tag: tags.comment, class: 'italic text-gray-400'},
            {tag: tags.propertyName, class: 'text-blue-800 dark:text-blue-300'},
          ])
        ),
        highlightSpecialChars(),
        yaml(),
        keymap.of([...defaultKeymap, ...historyKeymap, indentWithTab]),
      ],
      parent: this.host.nativeElement,
      dispatchTransactions: (trs, view) => {
        this.formCtrl()?.setValue(this.view.state.doc.toString());
        view.update(trs);
      },
    });
  }

  public ngOnDestroy(): void {
    this.destroyed$.next();
    this.destroyed$.complete();
    this.view.destroy();
  }
}
