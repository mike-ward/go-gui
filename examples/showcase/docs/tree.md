# Tree View

`tree` renders hierarchical data with expand/collapse behavior, keyboard navigation, optional virtualization, and lazy-loading hooks.

## Usage

```go
gui.Tree(gui.TreeCfg{
    ID:        "project-tree",
    IDFocus:   2001,
    IDScroll:  2002,
    Sizing:    gui.FillFit,
    MaxHeight: 240,
    OnSelect: func(nodeID string, _ *gui.Event, w *gui.Window) {
        gui.State[AppState](w).SelectedNode = nodeID
    },
    OnLazyLoad: func(treeID, nodeID string, w *gui.Window) {
        // Load children for nodeID, store them in app state, then refresh.
    },
    Nodes: []gui.TreeNodeCfg{
        {
            ID:   "src",
            Text: "src",
            Icon: gui.IconFolder,
            Nodes: []gui.TreeNodeCfg{
                {Text: "main.go"},
                {Text: "view_tree.go"},
            },
        },
        {ID: "remote", Text: "remote", Lazy: true},
    },
})
```

## Core Types

- `TreeCfg` controls sizing, styling, focus/scroll IDs, callbacks, and accessibility metadata.
- `TreeNodeCfg` describes each node's `ID`, `Text`, `Icon`, `Lazy`, nested `Nodes`, and optional text styles.

`TreeNodeCfg.ID` falls back to `Text` when omitted. Node IDs should be unique within a single tree.

## Behavior

- Click a row to focus it, toggle expansion for folders, and invoke `OnSelect`.
- Expand a lazy node with no loaded children to invoke `OnLazyLoad(treeID, nodeID, w)` once per loading cycle.
- Use `Up` / `Down` to move focus.
- Use `Left` to collapse and `Right` to expand.
- Use `Home` / `End` to jump to the first or last visible row.
- Use `Enter` or `Space` to invoke `OnSelect` for the focused row.

## Virtualization

Tree virtualization is enabled when:

- `IDScroll > 0`
- `Height` or `MaxHeight` is bounded

The widget flattens visible rows and uses spacer rows above and below the rendered window, similar to the existing list virtualization model.

## Accessibility

- The root uses the tree accessibility role.
- Rows use the tree item role.
- Expanded rows expose the expanded accessibility state.

## Showcase

The showcase includes:

- A basic nested tree
- A virtualized scrolling tree
- A lazy-loading tree backed by queued async updates
