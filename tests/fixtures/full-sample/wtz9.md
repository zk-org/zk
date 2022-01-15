# Use small `Hashable` items with diffable data sources

If `apply()` is too slow with a diffable data source, it's probably because the items take too long to be hashed. A best practice is to hash only the properties that are actually used for display in the view.

:programming:swift:ios:
