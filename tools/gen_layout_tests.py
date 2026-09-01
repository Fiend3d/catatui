"""Convert ratatui's rstest layout case tables into Go table tests.

These cases are the best available fidelity check for the layout port, so they
are translated mechanically rather than retyped. Run from the module root; the
output is layout_cases_test.go.
"""
import io
import re
import sys

SRC = '_ref/ratatui/ratatui-core/src/layout/layout.rs'
TAB = '\t'


def split_args(s):
    """Split a comma-separated argument list, respecting nesting and strings."""
    out, depth, cur, in_str = [], 0, '', False
    i = 0
    while i < len(s):
        c = s[i]
        if in_str:
            cur += c
            if c == '\\':
                cur += s[i + 1]
                i += 2
                continue
            if c == '"':
                in_str = False
        elif c == '"':
            in_str = True
            cur += c
        elif c in '([{':
            depth += 1
            cur += c
        elif c in ')]}':
            depth -= 1
            cur += c
        elif c == ',' and depth == 0:
            out.append(cur.strip())
            cur = ''
        else:
            cur += c
        i += 1
    if cur.strip():
        out.append(cur.strip())
    return out


def strip_comment(s):
    """Drop a trailing // comment that is not inside a string."""
    out, in_str, i = '', False, 0
    while i < len(s):
        if s[i] == '"' and (i == 0 or s[i - 1] != '\\'):
            in_str = not in_str
        if not in_str and s[i:i + 2] == '//':
            break
        out += s[i]
        i += 1
    return out.strip()


def collect_groups():
    """Return {fn_name: [case_arg_string, ...]} for the split test module."""
    src = io.open(SRC, encoding='utf-8').read()
    body = src[src.index('    mod split {'):]
    lines = [strip_comment(l.strip()) for l in body.split('\n')]
    groups, cur, i = {}, [], 0
    while i < len(lines):
        s = lines[i]
        if s.startswith('#[case') and '(' in s:
            buf = s
            while buf.count('(') > buf.count(')') and i + 1 < len(lines):
                i += 1
                buf += ' ' + lines[i]
            cur.append(buf[buf.index('(') + 1:buf.rindex(')')])
        elif s.startswith('fn ') and cur:
            groups[s[3:s.index('(')]] = cur
            cur = []
        elif s.startswith('#[rstest]'):
            cur = []
        i += 1
    return groups


FLEX = {
    'Flex::Legacy': 'FlexLegacy', 'Flex::Start': 'FlexStart', 'Flex::End': 'FlexEnd',
    'Flex::Center': 'FlexCenter', 'Flex::SpaceBetween': 'FlexSpaceBetween',
    'Flex::SpaceEvenly': 'FlexSpaceEvenly', 'Flex::SpaceAround': 'FlexSpaceAround',
}


def rust_ints(s):
    return s.replace('u16::MAX', '65535').replace('u32::MAX', '4294967295')


def go_constraints(s):
    s = rust_ints(s.strip())
    s = re.sub(r'^&?vec!\[', '[', s)
    s = re.sub(r'^&\[', '[', s).strip()
    assert s.startswith('[') and s.endswith(']'), s
    items = [re.sub(r'\s+', ' ', x) for x in split_args(s[1:-1])]
    return '[]Constraint{' + ', '.join(items) + '}'


def go_ranges(s):
    s = rust_ints(re.sub(r'^&?vec!\[', '[', s.strip()))
    out = []
    for it in split_args(s[1:-1]):
        a, b = it.split('..')
        out.append('{%s, %s}' % (a.strip(), b.strip()))
    return '[]rng{' + ', '.join(out) + '}'


def go_pairs(s):
    s = rust_ints(re.sub(r'^&?vec!\[', '[', s.strip()))
    out = []
    for it in split_args(s[1:-1]):
        it = it.strip()
        parts = split_args(it[1:-1])
        assert len(parts) == 2, (it, s)
        out.append('{%s, %s}' % (parts[0].strip(), parts[1].strip()))
    return '[]pair{' + ', '.join(out) + '}'


def go_u16s(s):
    s = rust_ints(re.sub(r'^&?vec!\[', '[', s.strip()))
    return '[]uint16{' + ', '.join(x.strip() for x in split_args(s[1:-1])) + '}'


groups = collect_groups()
out = []


def W(line=''):
    out.append(line)


def table(varname, comment, fields):
    for c in comment:
        W('// ' + c)
    W('var %s = []struct {' % varname)
    for fname, ftype in fields:
        W('%s%-11s %s' % (TAB, fname, ftype))
    W('}{')


W('// Code generated from ratatui-core/src/layout/layout.rs @ ratatui-v0.30.2 by')
W('// scratch_gen.py. DO NOT EDIT BY HAND.')
W('//')
W("// These are ratatui's own layout case tables, translated mechanically. They are")
W('// the fidelity gate for the layout port: if the Cassowary solver, the strength')
W('// ladder or the flex handling drifts from ratatui, these fail.')
W()
W('package catatui')
W()

counts = {}

# --- the `letters` harness -------------------------------------------------
LETTERS = ['length', 'max', 'min', 'percentage', 'percentage_start',
           'percentage_spacebetween', 'ratio', 'ratio_start', 'ratio_spacebetween']
table('lettersCases', [
    "lettersCases are driven by ratatui's `letters` helper: split a one-row area",
    'of the given width, fill each resulting segment with a repeated letter, and',
    'compare the rendered row against the expected string.',
], [('name', 'string'), ('flex', 'Flex'), ('width', 'uint16'),
    ('constraints', '[]Constraint'), ('expected', 'string')])
n = 0
for fn in LETTERS:
    for case in groups.get(fn, []):
        args = split_args(strip_comment(case))
        if len(args) != 4:
            print('SKIP %s: %r' % (fn, case), file=sys.stderr)
            continue
        flex, width, cons, exp = args
        W('%s{"%s", %s, %s, %s, %s},' % (TAB, fn, FLEX[flex.strip()], width.strip(),
                                         go_constraints(cons), exp.strip()))
        n += 1
W('}')
W()
counts['letters'] = n

# --- (constraints, ranges) at a fixed flex ---------------------------------
RANGES = {
    'constraint_length': (100, 'FlexLegacy'),
    'length_is_higher_priority': (100, 'FlexLegacy'),
    'fill': (100, 'FlexLegacy'),
    'percentage_parameterized': (100, 'FlexLegacy'),
    'min_max': (100, 'FlexLegacy'),
    'fixed_with_50_width': (50, 'FlexLegacy'),
}
table('rangeCases', ['rangeCases check the left..right span of each segment.'],
      [('name', 'string'), ('width', 'uint16'), ('flex', 'Flex'),
       ('constraints', '[]Constraint'), ('expected', '[]rng')])
n = 0
for fn, (width, flex) in RANGES.items():
    for case in groups.get(fn, []):
        args = split_args(strip_comment(case))
        if len(args) != 2:
            print('SKIP %s: %r' % (fn, case), file=sys.stderr)
            continue
        cons, exp = args
        W('%s{"%s", %d, %s, %s, %s},' % (TAB, fn, width, flex, go_constraints(cons), go_ranges(exp)))
        n += 1
W('}')
W()
counts['ranges'] = n

# --- flex_constraint: (constraints, ranges, flex) --------------------------
table('flexRangeCases', [
    'flexRangeCases check the span of each segment under an explicit flex mode,',
    'in a 100-column area.',
], [('name', 'string'), ('constraints', '[]Constraint'), ('flex', 'Flex'), ('expected', '[]rng')])
n = 0
for case in groups.get('flex_constraint', []):
    args = split_args(strip_comment(case))
    if len(args) != 3:
        print('SKIP flex_constraint: %r' % case, file=sys.stderr)
        continue
    cons, exp, flex = args
    W('%s{"flex_constraint", %s, %s, %s},' % (TAB, go_constraints(cons),
                                              FLEX[flex.strip()], go_ranges(exp)))
    n += 1
W('}')
W()
counts['flexranges'] = n

# --- (expected pairs, constraints, [flex], [spacing]) ----------------------
PAIRS = {
    'fill_vs_flex': ('ecf', None, '0'),
    'legacy_vs_default': ('ecf', None, '0'),
    'constraint_specification_tests_for_priority': ('ec', 'FlexLegacy', '0'),
    'flex_overlap': ('ecfs', None, None),
    'flex_spacing': ('ecfs', None, None),
    'fill_spacing': ('ecfs', None, None),
    'fill_overlap': ('ecfs', None, None),
    'constraint_specification_tests_for_priority_with_spacing': ('ecfs', None, None),
    'flex_spacing_lower_priority_than_user_spacing': ('ecfs', None, None),
}
table('pairCases', ['pairCases check the (x, width) of each segment.'],
      [('name', 'string'), ('constraints', '[]Constraint'), ('flex', 'Flex'),
       ('spacing', 'int'), ('expected', '[]pair')])
n = 0
for fn, (order, fixed_flex, fixed_spacing) in PAIRS.items():
    for case in groups.get(fn, []):
        args = split_args(strip_comment(case))
        if len(args) != len(order):
            print('SKIP %s: %r' % (fn, case), file=sys.stderr)
            continue
        exp, cons = args[0], args[1]
        flex = FLEX[args[2].strip()] if 'f' in order else fixed_flex
        spacing = args[3].strip() if 's' in order else fixed_spacing
        W('%s{"%s", %s, %s, %s, %s},' % (TAB, fn, go_constraints(cons), flex, spacing, go_pairs(exp)))
        n += 1
W('}')
W()
counts['pairs'] = n

# --- spacers ---------------------------------------------------------------
SPACERS = {
    'split_with_spacers_no_spacing': 'ecf',
    'split_with_spacers_and_spacing': 'ecfs',
    'split_with_spacers_and_overlap': 'ecfs',
    'split_with_spacers_and_too_much_spacing': 'ecfs',
}
table('spacerCases', [
    'spacerCases check the (x, width) of the gaps between segments, which widgets',
    'with collapsed borders draw into.',
], [('name', 'string'), ('constraints', '[]Constraint'), ('flex', 'Flex'),
    ('spacing', 'int'), ('expected', '[]pair')])
n = 0
for fn, order in SPACERS.items():
    for case in groups.get(fn, []):
        args = split_args(strip_comment(case))
        if len(args) != len(order):
            print('SKIP %s: %r' % (fn, case), file=sys.stderr)
            continue
        exp, cons = args[0], args[1]
        flex = FLEX[args[2].strip()]
        spacing = args[3].strip() if 's' in order else '0'
        W('%s{"%s", %s, %s, %s, %s},' % (TAB, fn, go_constraints(cons), flex, spacing, go_pairs(exp)))
        n += 1
W('}')
W()
counts['spacers'] = n

# --- widths across every non-legacy flex -----------------------------------
table('widthCases', [
    'widthCases check only the widths, and must hold across every non-Legacy flex',
    'mode.',
], [('name', 'string'), ('constraints', '[]Constraint'), ('expected', '[]uint16')])
n = 0
for case in groups.get('length_is_higher_priority_in_flex', []):
    args = split_args(strip_comment(case))
    if len(args) != 2:
        continue
    cons, exp = args
    W('%s{"length_is_higher_priority_in_flex", %s, %s},' % (TAB, go_constraints(cons), go_u16s(exp)))
    n += 1
W('}')
counts['widths'] = n

io.open('layout_cases_test.go', 'w', encoding='utf-8', newline='\n').write('\n'.join(out) + '\n')
print(' '.join('%s=%d' % kv for kv in counts.items()), 'total=%d' % sum(counts.values()))
print('now run: gofmt -w layout_cases_test.go')
