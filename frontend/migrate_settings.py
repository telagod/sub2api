#!/usr/bin/env python3
"""
Comprehensive SettingsView.vue migration to shadcn-vue.

Strategy:
  Pass 1: Line-by-line state machine for Card/CardHeader/CardContent structural changes
  Pass 2: Content-wide regex for multi-line <button class="btn ..."> -> <Button>
  Pass 3: Toggle -> Switch, import updates, button closing fixes
"""
import re

def main():
    filepath = 'src/views/admin/SettingsView.vue'
    with open(filepath, 'r') as f:
        content = f.read()

    # ==================== PASS 1: Card structure ====================
    content = pass1_cards(content)

    # ==================== PASS 2: Button replacements (multi-line) ====================
    content = pass2_buttons(content)

    # ==================== PASS 3: Toggle -> Switch ====================
    content = pass3_toggle_switch(content)

    # ==================== PASS 4: Import updates ====================
    content = pass4_imports(content)

    with open(filepath, 'w') as f:
        f.write(content)

    print_stats(content)


def pass1_cards(content: str) -> str:
    """Transform div.card to Card/CardHeader/CardContent structure."""
    lines = content.split('\n')
    result = []
    div_stack = []
    in_card_header = False
    pending_card_content = False
    card_desc_open = False

    i = 0
    while i < len(lines):
        line = lines[i]
        stripped = line.strip()
        skip_div_tracking = False  # Flag to skip generic div tracking for this line

        # CARD OPENING
        card_match = re.match(r'^(\s*)<div(\s+v-[^"]*"[^"]*")?\s*class="card">', line)
        if card_match:
            v_dir = card_match.group(2) or ''
            result.append(f'{card_match.group(1)}<Card{v_dir}>')
            div_stack.append('Card')
            i += 1
            pending_card_content = False

            # Detect header
            if i < len(lines):
                next_stripped = lines[i].strip()
                # Single-line header
                if next_stripped.startswith('<div') and 'border-b border-border' in next_stripped and next_stripped.endswith('>'):
                    h_indent = lines[i][:len(lines[i]) - len(lines[i].lstrip())]
                    result.append(f'{h_indent}<CardHeader>')
                    div_stack.append('CardHeader')
                    in_card_header = True
                    i += 1
                # Multi-line header: <div\n  class="... border-b ..."\n>
                elif next_stripped.startswith('<div') and '>' not in next_stripped:
                    if i + 1 < len(lines) and 'border-b border-border' in lines[i + 1]:
                        h_indent = lines[i][:len(lines[i]) - len(lines[i].lstrip())]
                        result.append(f'{h_indent}<CardHeader>')
                        div_stack.append('CardHeader')
                        in_card_header = True
                        i += 2
                        if i < len(lines) and lines[i].strip() == '>':
                            i += 1
                    else:
                        pending_card_content = True
                else:
                    pending_card_content = True
            continue  # Skip: Card opening already handled, no generic div tracking needed

        # PENDING CARD CONTENT
        if pending_card_content and stripped.startswith('<div'):
            content_match = re.match(r'^(\s*)<div class="((?:[^"]*\s)?p-\d+(?:\s[^"]*)?)">', line)
            if content_match:
                c_indent = content_match.group(1)
                classes = content_match.group(2)
                remaining = re.sub(r'\bp-\d+\b', '', classes).strip()
                if remaining:
                    result.append(f'{c_indent}<CardContent class="{remaining}">')
                else:
                    result.append(f'{c_indent}<CardContent>')
                div_stack.append('CardContent')
                pending_card_content = False
                skip_div_tracking = True  # Don't double-count this div
                i += 1
                continue
            else:
                pending_card_content = False

        # INSIDE CARD HEADER: transform h2/p
        if in_card_header:
            if '<h2 class="text-lg font-semibold text-foreground">' in line:
                line = line.replace(
                    '<h2 class="text-lg font-semibold text-foreground">',
                    '<CardTitle class="text-lg">'
                )
            if '</h2>' in line:
                line = line.replace('</h2>', '</CardTitle>')
            if '<p class="mt-1 text-sm text-muted-foreground">' in stripped:
                line = line.replace(
                    '<p class="mt-1 text-sm text-muted-foreground">',
                    '<CardDescription>'
                )
                card_desc_open = True
            if '</CardDescription>' in line:
                card_desc_open = False
            elif '</p>' in line and card_desc_open:
                line = line.replace('</p>', '</CardDescription>')
                card_desc_open = False

        # TRACK DIV OPENS - count all <div openings (handles multi-line divs)
        div_open_count = len(re.findall(r'<div\b', line))
        for _ in range(div_open_count):
            div_stack.append('div')

        # HANDLE DIV CLOSES
        close_count = line.count('</div>')
        for _ in range(close_count):
            if div_stack:
                tag = div_stack.pop()
                if tag == 'Card':
                    line = line.replace('</div>', '</Card>', 1)
                elif tag == 'CardHeader':
                    line = line.replace('</div>', '</CardHeader>', 1)
                    in_card_header = False
                    pending_card_content = True
                elif tag == 'CardContent':
                    line = line.replace('</div>', '</CardContent>', 1)

        # Toggle checkbox pattern
        if '<label class="toggle">' in stripped:
            toggle_lines = [line]
            i += 1
            while i < len(lines) and '</label>' not in lines[i]:
                toggle_lines.append(lines[i])
                i += 1
            if i < len(lines):
                toggle_lines.append(lines[i])
                i += 1
            toggle_text = '\n'.join(toggle_lines)
            m = re.search(r'v-model="([^"]+)"', toggle_text)
            if m:
                t_indent = re.match(r'^(\s*)', toggle_lines[0]).group(1)
                result.append(f'{t_indent}<Switch v-model="{m.group(1)}" />')
            else:
                result.extend(toggle_lines)
            continue

        result.append(line)
        i += 1

    return '\n'.join(result)


def pass2_buttons(content: str) -> str:
    """Replace multi-line <button ... class="btn ..."> with <Button>."""

    # Map of class patterns to replacement attributes
    replacements = [
        # btn-primary variants (most specific first)
        (r'class="btn btn-primary btn-sm flex-shrink-0"', 'size="sm" class="flex-shrink-0"'),
        (r'class="btn btn-primary btn-sm inline-flex items-center gap-1\.5"', 'size="sm" class="inline-flex items-center gap-1.5"'),
        (r'class="btn btn-primary btn-sm"', 'size="sm"'),
        (r'class="btn btn-primary"', ''),

        # btn-secondary variants
        (r'class="btn btn-secondary btn-sm text-red-400 hover:text-red-300"', 'variant="secondary" size="sm" class="text-red-400 hover:text-red-300"'),
        (r'class="btn btn-secondary btn-sm inline-flex items-center gap-1"', 'variant="secondary" size="sm" class="inline-flex items-center gap-1"'),
        (r'class="btn btn-secondary btn-sm whitespace-nowrap"', 'variant="secondary" size="sm" class="whitespace-nowrap"'),
        (r'class="btn btn-secondary btn-sm w-fit"', 'variant="secondary" size="sm" class="w-fit"'),
        (r'class="btn btn-secondary btn-sm"', 'variant="secondary" size="sm"'),
        (r'class="btn btn-secondary default-sub-delete-btn w-full text-red-400 hover:text-red-300"', 'variant="secondary" class="default-sub-delete-btn w-full text-red-400 hover:text-red-300"'),
        (r'class="btn btn-secondary w-full text-red-400 hover:text-red-300"', 'variant="secondary" class="w-full text-red-400 hover:text-red-300"'),
        (r'class="btn btn-secondary px-2"', 'variant="secondary" class="px-2"'),
        (r'class="btn btn-secondary"', 'variant="secondary"'),

        # btn-ghost variants
        (r'class="btn btn-ghost btn-xs text-red-400 hover:text-red-300"', 'variant="ghost" size="sm" class="text-red-400 hover:text-red-300 h-auto px-1.5 py-0.5"'),
        (r'class="btn btn-ghost btn-xs text-primary-600 dark:text-primary-400"', 'variant="ghost" size="sm" class="text-primary-600 dark:text-primary-400 h-auto px-1.5 py-0.5"'),
    ]

    # Attribute pattern: matches name="value" or name='value' or bare-name
    # Handles > inside quoted attribute values (e.g., v-if="x > 0")
    ATTR = r'(?:[\w@:.v-]+=(?:"[^"]*"|\'[^\']*\')|\w+)'

    for class_pattern, attrs in replacements:
        pattern = r'(<button\b)((?:\s+' + ATTR + r')*)\s*\n?\s*' + class_pattern
        if attrs:
            replacement = r'<Button\2\n                  ' + attrs
        else:
            replacement = r'<Button\2'

        content = re.sub(pattern, replacement, content, flags=re.DOTALL)

    return content


def pass3_toggle_switch(content: str) -> str:
    """Replace <Toggle> with <Switch>."""
    content = re.sub(r'<Toggle\b', '<Switch', content)
    content = content.replace('</Toggle>', '</Switch>')
    return content


def pass4_imports(content: str) -> str:
    """Update imports."""
    content = content.replace(
        'import Toggle from "@/components/common/Toggle.vue";',
        'import { Switch } from "@/components/ui/switch";'
    )
    content = content.replace(
        'import Icon from "@/components/icons/Icon.vue";',
        'import Icon from "@/components/icons/Icon.vue";\nimport { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";\nimport { Button } from "@/components/ui/button";'
    )

    # Fix </button> -> </Button>
    content = fix_button_closings(content)

    return content


def fix_button_closings(content: str) -> str:
    """Fix </button> that should be </Button>."""
    lines = content.split('\n')
    result = []
    btn_stack = []

    for line in lines:
        if re.search(r'<Button\b', line) and '/>' not in line and '</Button>' not in line:
            btn_stack.append(True)
        elif re.search(r'<button\b', line) and '/>' not in line and '</button>' not in line:
            btn_stack.append(False)

        if '</button>' in line:
            if btn_stack and btn_stack[-1]:
                line = line.replace('</button>', '</Button>', 1)
                btn_stack.pop()
            elif btn_stack:
                btn_stack.pop()

        result.append(line)

    return '\n'.join(result)


def print_stats(content: str):
    stats = [
        ('<Card opens', len(re.findall(r'<Card[\s>]', content))),
        ('</Card>', content.count('</Card>')),
        ('<CardHeader>', content.count('<CardHeader>')),
        ('</CardHeader>', content.count('</CardHeader>')),
        ('<CardContent', len(re.findall(r'<CardContent[\s>]', content))),
        ('</CardContent>', content.count('</CardContent>')),
        ('<CardTitle', content.count('<CardTitle')),
        ('</CardTitle>', content.count('</CardTitle>')),
        ('<CardDescription>', content.count('<CardDescription>')),
        ('</CardDescription>', content.count('</CardDescription>')),
        ('<Button opens', len(re.findall(r'<Button[\s\n]', content))),
        ('</Button>', content.count('</Button>')),
        ('<Switch', content.count('<Switch')),
        ('rem class="card"', content.count('class="card"')),
        ('rem class="btn ', content.count('class="btn ')),
        ('rem <Toggle', content.count('<Toggle')),
    ]
    for name, count in stats:
        print(f'  {name}: {count}')


if __name__ == '__main__':
    main()
