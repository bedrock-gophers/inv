# Inv
The Inv library allows the use of inventory menus, providing tools to create or send fake inventories.

## Creating an Inventory Menu
To create an inventory menu using the Inv library, follow these steps:

Place Fake Chest: use the inv.PlaceFakeChest function to place a fake chest, essential for the library to function properly.
```go
// Only run this once, after server start.
inv.PlaceFakeChest(myWorld, cube.Pos{x, y, z)
```

Create Inventory: Use the inventory.New function to create a new inventory with the desired size and callback function.

```go
in := inventory.New(27, func(slot int, before, after item.Stack) {})
```
Populate Inventory Slots: Iterate over the inventory slots and set items accordingly.

```go
for i := range in.Slots() {
    _ = in.SetItem(i, item.NewStack(block.StainedGlassPane{
        Colour: item.ColourRed(),
    }, 1))
}
```
Creating a Handler for Inventories
You can create a custom handler for inventories to handle inventory events. Here's an example:

```go
// InventoryHandler represents our custom inventory handler.
type InventoryHandler struct {
    inventory.NopHandler
}

// HandleTake ...
func (h InventoryHandler) HandleTake(ctx *event.Context, slot int, it item.Stack) {
    fmt.Println(slot)
}
```
## Handling Inventory Events
Once you've created a handler, you can associate it with an inventory to handle inventory events. Here's how:

```go
in.Handle(InventoryHandler{})
```
Sending the Menu to a Player
To display the inventory menu to a player, use the inv.ShowMenu function. Here's an example:

```go
// The 'p' variable represents the targeted player.
inv.ShowMenu(p, in, text.Colourf("<red>Test</red>"))
```
This code sends the inventory menu to the specified player with the provided title.
