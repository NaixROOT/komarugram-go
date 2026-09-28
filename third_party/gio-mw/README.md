# Gio - Material Widgets

This is a library containing various small [Gio](https://gioui.org/) widgets and examples.

## Dependencies

Built in [Go](https://go.dev/doc/install) with [Gio](https://gioui.org/doc/install) and one extra dependency:

* https://pkg.go.dev/golang.org/x/exp/shiny — for UI icons

## License

This project is licensed under the same terms as [Gio](https://gioui.org/) to ensure the best compatibility. Specifically, it is provided under the terms of the [UNLICENSE](https://opensource.org/license/Unlicense) or the [MIT license](https://opensource.org/license/MIT), as denoted by the following SPDX identifier:

SPDX-License-Identifier: Unlicense OR MIT

You may choose to use the project under either license.

## Examples

For an example implementation, run the Kitchen Sink example:

    go run ./examples/kitchen/

Or check it out in your browser:

* https://schnwalter.eu/wasm/kitchen/

Or check out these applications:

* https://git.sr.ht/~schnwalter/hypatia — a Gemini Protocol browser
* https://git.sr.ht/~schnwalter/enb — a way to change your EFI Next Boot target

## Material Components Status

| Component           | Status            |
|---------------------|-------------------|
| App bars            | ❌ Not implemented |
| Badges              | ✔️ Implemented    |
| Buttons             | ✔️ Implemented    |
| Buttons: FAB        | ❌ Not implemented |
| Buttons: group      | ❌ Not implemented |
| Buttons: icon       | ✔️ Implemented    |
| Buttons: segmented  | ❌ Not implemented |
| Buttons: split      | ❌ Not implemented |
| Cards               | ✔️ Implemented    |
| Carousel            | ❌ Not implemented |
| Checkbox            | ✔️ Implemented    |
| Chips               | ❌ Not implemented |
| Date pickers        | ❌ Not implemented |
| Dialogs             | ✔️ Implemented    |
| Divider             | ✔️ Implemented    |
| Lists               | ❌ Not implemented |
| Loading indicators  | ❌ Not implemented |
| Menus               | ❌ Not implemented |
| Navigation bar      | ❌ Not implemented |
| Navigation rail     | ✔️ Implemented    |
| Progress indicators | ✔️ Implemented    |
| Radio button        | ✔️ Implemented    |
| Search              | ✔️ Implemented    |
| Sheets              | ✔️ Implemented    |
| Sliders             | ✔️ Implemented    |
| Snackbar            | ✔️ Implemented    |
| Switch              | ✔️ Implemented    |
| Tabs                | ✔️ Implemented    |
| Text fields         | ✔️ Implemented    |
| Time pickers        | ❌ Not implemented |
| Toolbars            | ❌ Not implemented |
| Tooltips            | ✔️ Implemented    |

## Material Features Status

| Feature           | Status            |
|-------------------|-------------------|
| Color schemes     | ✔️ Implemented    |
| Component themes  | ✔️ Implemented    |
| Dark mode         | ✔️ Implemented    |
| Elevation         | ✔️ Implemented    |
| Icons             | ✔️ Implemented    |
| Light mode        | ✔️ Implemented    |
| Motion system     | ❌ Not implemented |
| Ripple effects    | ❌ Not implemented |
| Shape system      | ❌ Not implemented |
| Shaped corners    | ✔️ Implemented    |
| Typography system | ✔️ Implemented    |

## Development

To use it within your own project while developing it, you'll have to clone the project and use it as a [workspace](https://go.dev/doc/tutorial/workspaces) module.

Create a `go.work` file in the root of your project and point it to the clone of this repo:

    go 1.24.4
    use (
        .
        ../schnwalter--gio-mw
    )

For testing local Gio changes:

    go 1.24.4
    use (
        .
        ../eliasnaur--gio
    )

## Update

To update gio-mw to the latest version, run:

    go get -u gio-mw@latest

## More Info

To add new color schemes, see [README](./defaults/schemes/README.md)
