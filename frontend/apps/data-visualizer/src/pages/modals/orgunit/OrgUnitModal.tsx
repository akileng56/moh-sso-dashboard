import { Modal, MultiSelect } from "@carbon/react";
import { useEffect, useState } from "react";
import { useGetHierarchyQuery } from "./org-unit.ts";
import { levelOfCareOptions, ownershipOptions } from "../../Contants.tsx"

const normalizeSearchTerm = (value = "") => value.trim().toLowerCase();

const getUniqueChildren = (childrenObj = {}) =>
  Object.values(childrenObj).filter(
    (child: any, index, self) => index === self.findIndex((c: any) => c.id === child.id),
  );

const filterNodeBySearch = (node: any, searchTerm: string) => {
  const normalizedSearch = normalizeSearchTerm(searchTerm);

  if (!normalizedSearch) {
    return { node, expandedIds: new Set() };
  }

  const matchesNode = node.name?.toLowerCase().includes(normalizedSearch);
  const uniqueChildren = getUniqueChildren(node.children);
  const expandedIds = new Set<string>();

  if (matchesNode) {
    if (uniqueChildren.length > 0) {
      expandedIds.add(node.uid);
    }
    return { node, expandedIds };
  }

  const filteredChildren = uniqueChildren
    .map((child: any) => filterNodeBySearch(child, normalizedSearch))
    .filter((result) => result.node);

  if (filteredChildren.length === 0) {
    return { node: null, expandedIds };
  }

  filteredChildren.forEach((result) => {
    result.expandedIds.forEach((id) => expandedIds.add(id));
  });
  expandedIds.add(node.uid);

  return {
    node: {
      ...node,
      children: Object.fromEntries(filteredChildren.map((result) => [result.node.uid, result.node])),
    },
    expandedIds,
  };
};

function TreeNode({ node, level = 0, selectedUnits, onToggle, onExpand, expandedNodes }) {
  const childrenObj = node.children || {};
  const uniqueChildren = getUniqueChildren(childrenObj);

  const hasChildren = uniqueChildren.length > 0;
  const isExpanded = expandedNodes.has(node.uid);
  const isSelected = selectedUnits.has(node.uid);
  const childCount = uniqueChildren.length;

  const handleToggle = () => {
    onToggle(node.uid);
  };

  const handleExpand = () => {
    if (hasChildren) {
      onExpand(node.uid);
    }
  };

  const getIcon = () => {
    if (hasChildren) {
      return isExpanded ? "fas fa-chevron-down" : "fas fa-chevron-right";
    }
    return "";
  };

  const indent = level * 18;
  const childrenPadding = indent + 18;

  return (
    <div>
      <div
        className="d-flex align-items-center p-1"
        style={{
          paddingLeft: `${indent}px`,
        }}
      >
        <button
          className="btn btn-sm btn-link p-0 me-2"
          style={{ width: "16px", visibility: hasChildren ? "visible" : "hidden" }}
          onClick={handleExpand}
          disabled={!hasChildren}
        >
          {hasChildren && <i className={getIcon()}></i>}
        </button>
        <input
          type="checkbox"
          className="form-check-input me-2"
          checked={isSelected}
          onChange={handleToggle}
        />
        <span className="d-flex">
          {node.name}
          {childCount > 0 && <span className="text-muted"> ({childCount})</span>}
        </span>
      </div>

      {hasChildren && isExpanded && (
        <div
          style={{
            paddingLeft: `${childrenPadding}px`,
            borderLeft: "1px solid #eee",
          }}
        >
          {uniqueChildren.map((child: any) => (
            <TreeNode
              key={child.uid}
              node={child}
              level={level + 1}
              selectedUnits={selectedUnits}
              onToggle={onToggle}
              onExpand={onExpand}
              expandedNodes={expandedNodes}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export default function OrgUnitModal({
  onClose,
  selected,
  selectedLevelOfCare = [],
  selectedOwnership = [],
  onSave,
}) {
  const [selectedUnits, setSelectedUnits] = useState(new Set(selected ?? []));
  const [levelOfCare, setLevelOfCare] = useState(selectedLevelOfCare);
  const [ownership, setOwnership] = useState(selectedOwnership);
  const [expandedNodes, setExpandedNodes] = useState(new Set());
  const [orgUnits, setOrgUnits] = useState<any>({});
  const [searchTerm, setSearchTerm] = useState("");
  const { data: hierarchyData, isLoading, error } = useGetHierarchyQuery();

  useEffect(() => {
    if (!isLoading) {
      setOrgUnits(hierarchyData);

      const expandIds = new Set();
      const rootKeys = Object.keys(hierarchyData);

      if (rootKeys.length > 0) {
        const firstRootNode = hierarchyData[rootKeys[0]];
        expandIds.add(firstRootNode.uid);

        const firstChildKeys = Object.keys(firstRootNode.children || {});
        if (firstChildKeys.length > 0) {
          const firstChildNode = firstRootNode.children[firstChildKeys[0]];
          expandIds.add(firstChildNode.uid);
        }
      }
      setExpandedNodes(expandIds);
    }
  }, [hierarchyData, isLoading]);

  useEffect(() => {
    setSelectedUnits(new Set(selected ?? []));
  }, [selected]);

  useEffect(() => {
    setLevelOfCare(selectedLevelOfCare ?? []);
  }, [selectedLevelOfCare]);

  useEffect(() => {
    setOwnership(selectedOwnership ?? []);
  }, [selectedOwnership]);

  const toggleUnit = (unitUid) => {
    const newSelected = new Set(selectedUnits);
    if (newSelected.has(unitUid)) {
      newSelected.delete(unitUid);
    } else {
      newSelected.add(unitUid);
    }
    setSelectedUnits(newSelected);
  };

  const toggleExpanded = (unitUid) => {
    const newExpanded = new Set(expandedNodes);
    if (newExpanded.has(unitUid)) {
      newExpanded.delete(unitUid);
    } else {
      newExpanded.add(unitUid);
    }
    setExpandedNodes(newExpanded);
  };

  const deselectAll = () => {
    setSelectedUnits(new Set());
  };

  const save = () => {
    onSave(Array.from(selectedUnits), levelOfCare, ownership);
    onClose();
  };

  const selectedCount = selectedUnits.size;
  const normalizedSearchTerm = normalizeSearchTerm(searchTerm);
  const filteredResults = Object.values(orgUnits).map((unit: any) => filterNodeBySearch(unit, normalizedSearchTerm));
  const visibleUnits = filteredResults
    .map((result) => result.node)
    .filter(Boolean);
  const searchExpandedNodes = filteredResults.reduce((acc, result) => {
    result.expandedIds.forEach((id) => acc.add(id));
    return acc;
  }, new Set());
  const effectiveExpandedNodes = normalizedSearchTerm
    ? new Set([...expandedNodes, ...searchExpandedNodes])
    : expandedNodes;

  return (
    <>
      <Modal
        open
        size="md"
        preventCloseOnClickOutside={true}
        hasScrollingContent={true}
        modalHeading="Organisation Units"
        secondaryButtonText="Hide"
        primaryButtonText="Update"
        onRequestClose={onClose}
        onRequestSubmit={save}
      >
        <div className="border" style={{ height: "400px", overflowY: "auto" }}>
          <div
            className="p-2 border-bottom bg-white"
            style={{ position: "sticky", top: 0, zIndex: 1 }}
          >
            <input
              type="search"
              className="form-control"
              placeholder="Search organisation unit"
              value={searchTerm}
              onChange={(event) => setSearchTerm(event.target.value)}
            />
          </div>
          {isLoading ? (
            <div className="d-flex justify-content-center align-items-center h-100">
              <div className="spinner-border text-primary" role="status">
                <span className="visually-hidden">Loading...</span>
              </div>
            </div>
          ) : error ? (
            <div className="d-flex justify-content-center align-items-center h-100">
              <div className="text-center text-danger">
                <i className="fas fa-exclamation-triangle fa-2x mb-2"></i>
                <div>
                  {"status" in error
                    ? `Error ${error.status}: ${JSON.stringify(error.data)}`
                    : error.message}
                </div>
              </div>
            </div>
          ) : Object.keys(orgUnits).length === 0 ? (
            <div className="d-flex justify-content-center align-items-center h-100">
              <div className="text-center text-muted">
                <i className="fas fa-folder-open fa-2x mb-2"></i>
                <div>No organizational units found</div>
              </div>
            </div>
          ) : visibleUnits.length === 0 ? (
            <div className="d-flex justify-content-center align-items-center h-100">
              <div className="text-center text-muted">
                <i className="fas fa-search fa-2x mb-2"></i>
                <div>No organization units match "{searchTerm.trim()}"</div>
              </div>
            </div>
          ) : (
            visibleUnits.map((unit: any) => (
              <TreeNode
                key={unit.uid}
                node={unit}
                selectedUnits={selectedUnits}
                onToggle={toggleUnit}
                onExpand={toggleExpanded}
                expandedNodes={effectiveExpandedNodes}
              />
            ))
          )}
        </div>

        <div className="mt-3 d-flex justify-content-between align-items-center">
          <span className="text-muted">
            {selectedCount} selected
            {selectedCount > 0 && (
              <button className="btn btn-link p-0 ms-2 text-decoration-none" onClick={deselectAll}>
                - Deselect all
              </button>
            )}
          </span>
        </div>

        <div className="row mb-3">
          <div className="col-md-6">
            <MultiSelect
              id="level-of-care-select"
              titleText="Level of care"
              label="Select level of care"
              items={levelOfCareOptions}
              itemToString={(item) => item?.label ?? ""}
              selectedItems={levelOfCareOptions.filter((option) => levelOfCare.includes(option.id))}
              onChange={({ selectedItems }) => {
                setLevelOfCare((selectedItems ?? []).map((item) => item.id));
              }}
            />
          </div>
          <div className="col-md-6">
            <MultiSelect
              id="ownership-select"
              titleText="Ownership"
              label="Select ownership"
              items={ownershipOptions}
              itemToString={(item) => item?.label ?? ""}
              selectedItems={ownershipOptions.filter((option) => ownership.includes(option.id))}
              onChange={({ selectedItems }) => {
                setOwnership((selectedItems ?? []).map((item) => item.id));
              }}
            />
          </div>
        </div>
      </Modal>
    </>
  );
}
