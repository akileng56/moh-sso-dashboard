package admin_units

// BuildHierarchyTree builds the existing display hierarchy from caller-supplied
// rows. Hosts must authenticate, authorize and scope those rows before calling.
// It preserves fallback matching, ordering, deduplication and child nullability.
func BuildHierarchyTree(rows []OrgUnit) []TreeNode {
	return buildHierarchyTree(rows)
}
