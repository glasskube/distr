import {Component, ElementRef, forwardRef, inject, OnDestroy, OnInit} from '@angular/core';
import {ControlValueAccessor, NG_VALUE_ACCESSOR} from '@angular/forms';
import {defaultKeymap, history, historyKeymap, indentWithTab} from '@codemirror/commands';
import {yaml} from '@codemirror/lang-yaml';
import {HighlightStyle, indentOnInput, syntaxHighlighting} from '@codemirror/language';
import {EditorView, highlightSpecialChars, keymap} from '@codemirror/view';
import {tags} from '@lezer/highlight';
import {Subject} from 'rxjs';

@Component({
  selector: 'app-yaml-editor',
  template: '',
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => YamlEditorComponent),
      multi: true,
    },
  ],
})
export class YamlEditorComponent implements OnInit, OnDestroy, ControlValueAccessor {
  private readonly host = inject(ElementRef);
  private view!: EditorView;
  private readonly destroyed$ = new Subject<void>();

  disabled = false;

  public ngOnInit(): void {
    this.view = new EditorView({
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
        EditorView.updateListener.of((update) => {
          this.onTouched();
          if (update.docChanged) {
            this.onChange(this.view.state.doc.toString());
          }
        }),
      ],
      parent: this.host.nativeElement,
    });
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
    this.view.destroy();
  }

  writeValue(value: string) {
    const tr = this.view.state.update({changes: {from: 0, to: this.view.state.doc.length, insert: value ?? ''}});
    this.view.dispatch(tr);
  }

  registerOnChange(fn: (value: string) => void) {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void) {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean) {
    // TODO implement
    if (isDisabled) {
      console.error('setDisabledState not implemented yet');
    }
  }

  private onChange = (_: any) => {};

  private onTouched = () => {};
}
