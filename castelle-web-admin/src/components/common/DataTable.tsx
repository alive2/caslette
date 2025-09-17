import React, { useState, useMemo, useRef, useEffect } from "react";
import { createPortal } from "react-dom";
import {
  FaSort,
  FaSortUp,
  FaSortDown,
  FaChevronLeft,
  FaChevronRight,
  FaSearch,
  FaFilter,
  FaEllipsisV,
  FaChevronDown,
} from "react-icons/fa";
import { cosmic } from "../../styles/cosmic-theme";

export interface DataTableColumn<T = Record<string, unknown>> {
  key: string;
  title: string;
  sortable?: boolean;
  render?: (value: unknown, row: T, index: number) => React.ReactNode;
  className?: string;
  width?: string;
  mobileLabel?: string; // Custom label for mobile card view
  hiddenOnMobile?: boolean; // Hide this column on mobile
}

export interface DataTableAction<T = Record<string, unknown>> {
  label: string;
  icon?: React.ReactNode;
  onClick: (row: T, index: number) => void;
  className?: string;
  condition?: (row: T) => boolean; // Show action only if condition is true
}

export interface DataTableProps<T = Record<string, unknown>> {
  data: T[];
  columns: DataTableColumn<T>[];
  actions?: DataTableAction<T>[];
  loading?: boolean;
  pagination?: {
    page: number;
    limit: number;
    total: number;
    onPageChange: (page: number) => void;
  };
  searchable?: boolean;
  filterable?: boolean;
  onSearch?: (query: string) => void;
  onFilter?: (filters: Record<string, unknown>) => void;
  emptyMessage?: string;
  className?: string;
  cardKeyExtractor?: (row: T) => string | number;
  cardTitleExtractor?: (row: T) => string;
  cardSubtitleExtractor?: (row: T) => string;
  mobileBreakpoint?: string; // Tailwind breakpoint (default: 'md')
  useDropdownActions?: boolean; // Whether to use dropdown for actions in desktop view
}

export default function DataTable<T = Record<string, unknown>>({
  data,
  columns,
  actions = [],
  loading = false,
  pagination,
  searchable = false,
  filterable = false,
  onSearch,
  emptyMessage = "No data available",
  className = "",
  cardKeyExtractor = (row: T) =>
    ((row as Record<string, unknown>).id as string) || Math.random().toString(),
  cardTitleExtractor,
  cardSubtitleExtractor,
  mobileBreakpoint = "md",
  useDropdownActions = false,
}: DataTableProps<T>) {
  const [sortColumn, setSortColumn] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<"asc" | "desc">("asc");
  const [searchQuery, setSearchQuery] = useState("");
  const [showFilters, setShowFilters] = useState(false);
  const [openDropdown, setOpenDropdown] = useState<string | null>(null);
  const [dropdownPosition, setDropdownPosition] = useState<{
    top: number;
    left: number;
  } | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const buttonRefs = useRef<{ [key: string]: HTMLButtonElement }>({});

  // Handle dropdown toggle with portal positioning
  const handleDropdownToggle = (rowId: string) => {
    if (openDropdown === rowId) {
      setOpenDropdown(null);
      setDropdownPosition(null);
    } else {
      const button = buttonRefs.current[rowId];
      if (button) {
        const rect = button.getBoundingClientRect();
        const dropdownWidth = 192; // min-w-48 = 192px

        // Calculate position for dropdown
        let left = rect.right - dropdownWidth;

        // Keep dropdown within viewport
        if (left < 8) {
          left = 8;
        }
        if (left + dropdownWidth > window.innerWidth - 8) {
          left = window.innerWidth - dropdownWidth - 8;
        }

        setDropdownPosition({
          top: rect.bottom + window.scrollY + 4,
          left: left,
        });
        setOpenDropdown(rowId);
      }
    }
  };

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setOpenDropdown(null);
        setDropdownPosition(null);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  // Handle sorting
  const handleSort = (columnKey: string) => {
    const column = columns.find((col) => col.key === columnKey);
    if (!column?.sortable) return;

    if (sortColumn === columnKey) {
      setSortDirection(sortDirection === "asc" ? "desc" : "asc");
    } else {
      setSortColumn(columnKey);
      setSortDirection("asc");
    }
  };

  // Sort data
  const sortedData = useMemo(() => {
    if (!sortColumn) return data;

    return [...data].sort((a, b) => {
      const aValue = (a as Record<string, unknown>)[sortColumn];
      const bValue = (b as Record<string, unknown>)[sortColumn];

      if (aValue === bValue) return 0;

      let comparison = 0;
      if (typeof aValue === "string" && typeof bValue === "string") {
        comparison = aValue.localeCompare(bValue);
      } else if (typeof aValue === "number" && typeof bValue === "number") {
        comparison = aValue - bValue;
      } else if (aValue instanceof Date && bValue instanceof Date) {
        comparison = aValue.getTime() - bValue.getTime();
      } else {
        comparison = String(aValue).localeCompare(String(bValue));
      }

      return sortDirection === "desc" ? -comparison : comparison;
    });
  }, [data, sortColumn, sortDirection]);

  // Handle search
  const handleSearchChange = (query: string) => {
    setSearchQuery(query);
    onSearch?.(query);
  };

  // Render sort icon
  const renderSortIcon = (columnKey: string) => {
    const column = columns.find((col) => col.key === columnKey);
    if (!column?.sortable) return null;

    if (sortColumn !== columnKey) {
      return <FaSort className="text-gray-400 text-xs" />;
    }

    return sortDirection === "asc" ? (
      <FaSortUp className="text-purple-400 text-xs" />
    ) : (
      <FaSortDown className="text-purple-400 text-xs" />
    );
  };

  // Render table view (desktop)
  const renderTable = () => {
    return (
      <div className="overflow-x-auto overflow-y-visible">
        <table className="min-w-full divide-y divide-white/10 relative">
          <thead className="bg-gradient-to-r from-purple-800/50 to-blue-800/50">
            <tr>
              {columns.map((column) => (
                <th
                  key={column.key}
                  className={`px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider cursor-pointer hover:text-white transition-colors ${
                    column.className || ""
                  }`}
                  style={{ width: column.width }}
                  onClick={() => handleSort(column.key)}
                >
                  <div className="flex items-center space-x-2">
                    <span>{column.title}</span>
                    {renderSortIcon(column.key)}
                  </div>
                </th>
              ))}
              {actions.length > 0 && (
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-300 uppercase tracking-wider">
                  Actions
                </th>
              )}
            </tr>
          </thead>
          <tbody className="divide-y divide-white/10">
            {sortedData.map((row, index) => (
              <tr
                key={cardKeyExtractor(row)}
                className="hover:bg-white/5 transition-colors"
              >
                {columns.map((column) => (
                  <td
                    key={column.key}
                    className={`px-6 py-4 whitespace-nowrap text-sm text-white ${
                      column.className || ""
                    }`}
                  >
                    {column.render
                      ? column.render(
                          (row as Record<string, unknown>)[column.key],
                          row,
                          index
                        )
                      : String(
                          (row as Record<string, unknown>)[column.key] || ""
                        )}
                  </td>
                ))}
                {actions.length > 0 && (
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    {useDropdownActions ? (
                      <div
                        className="relative inline-block text-right"
                        ref={dropdownRef}
                      >
                        <button
                          ref={(el) => {
                            if (el)
                              buttonRefs.current[
                                cardKeyExtractor(row).toString()
                              ] = el;
                          }}
                          onClick={() =>
                            handleDropdownToggle(
                              cardKeyExtractor(row).toString()
                            )
                          }
                          className={`${cosmic.button.ghost} p-2 rounded-full hover:bg-white/10 transition-colors ml-auto flex items-center justify-center`}
                          title="Actions"
                        >
                          <FaEllipsisV className="w-3 h-3" />
                        </button>
                        {openDropdown === cardKeyExtractor(row) &&
                          dropdownPosition &&
                          createPortal(
                            <div
                              className="fixed bg-gray-900/95 backdrop-blur-sm border border-gray-600/50 rounded-lg shadow-xl z-[9999] min-w-48 max-w-xs"
                              style={{
                                top: `${dropdownPosition.top}px`,
                                left: `${dropdownPosition.left}px`,
                              }}
                              ref={dropdownRef}
                            >
                              {actions.map((action, actionIndex) => {
                                if (
                                  action.condition &&
                                  !action.condition(row)
                                ) {
                                  return null;
                                }
                                return (
                                  <button
                                    key={actionIndex}
                                    onClick={() => {
                                      action.onClick(row, index);
                                      setOpenDropdown(null);
                                      setDropdownPosition(null);
                                    }}
                                    className="w-full px-4 py-3 text-left text-sm text-white hover:bg-white/10 flex items-center space-x-3 first:rounded-t-lg last:rounded-b-lg transition-colors"
                                  >
                                    <span className="text-purple-400 w-4 h-4 flex items-center justify-center">
                                      {action.icon}
                                    </span>
                                    <span className="flex-1">
                                      {action.label}
                                    </span>
                                  </button>
                                );
                              })}
                            </div>,
                            document.body
                          )}
                      </div>
                    ) : (
                      <div className="flex justify-end space-x-2">
                        {actions.map((action, actionIndex) => {
                          if (action.condition && !action.condition(row)) {
                            return null;
                          }
                          return (
                            <button
                              key={actionIndex}
                              onClick={() => action.onClick(row, index)}
                              className={`${
                                action.className || cosmic.button.ghost
                              } text-xs flex items-center space-x-1`}
                              title={action.label}
                            >
                              {action.icon}
                              <span className="hidden sm:inline">
                                {action.label}
                              </span>
                            </button>
                          );
                        })}
                      </div>
                    )}
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  };

  // Render card view (mobile)
  const renderCards = () => (
    <div className="space-y-4">
      {sortedData.map((row, index) => (
        <div
          key={cardKeyExtractor(row)}
          className={`${cosmic.cardElevated} p-4`}
        >
          {/* Card Header */}
          {(cardTitleExtractor || cardSubtitleExtractor) && (
            <div className="border-b border-white/10 pb-3 mb-3">
              {cardTitleExtractor && (
                <h3
                  className={`text-lg font-medium text-white ${
                    /^[+\-0-9]/.test(cardTitleExtractor(row))
                      ? "text-right"
                      : ""
                  }`}
                >
                  {cardTitleExtractor(row)}
                </h3>
              )}
              {cardSubtitleExtractor && (
                <p className="text-sm text-gray-400 mt-1">
                  {cardSubtitleExtractor(row)}
                </p>
              )}
            </div>
          )}

          {/* Card Content */}
          <div className="space-y-3">
            {columns
              .filter((column) => !column.hiddenOnMobile)
              .map((column) => (
                <div
                  key={column.key}
                  className="flex justify-between items-start"
                >
                  <span className="text-sm font-medium text-gray-300">
                    {column.mobileLabel || column.title}:
                  </span>
                  <span className="text-sm text-white text-right flex-1 ml-3">
                    {column.render
                      ? column.render(
                          (row as Record<string, unknown>)[column.key],
                          row,
                          index
                        )
                      : String(
                          (row as Record<string, unknown>)[column.key] || ""
                        )}
                  </span>
                </div>
              ))}
          </div>

          {/* Card Actions */}
          {actions.length > 0 && (
            <div className="border-t border-white/10 pt-3 mt-3">
              <div className="flex flex-wrap gap-2 justify-end">
                {/* Show only first 2 actions directly */}
                {actions.slice(0, 2).map((action, actionIndex) => {
                  if (action.condition && !action.condition(row)) {
                    return null;
                  }
                  return (
                    <button
                      key={actionIndex}
                      onClick={() => action.onClick(row, index)}
                      className={`${
                        action.className || cosmic.button.ghost
                      } text-xs flex items-center space-x-1`}
                    >
                      {action.icon}
                      <span>{action.label}</span>
                    </button>
                  );
                })}

                {/* Show dropdown for remaining actions if there are more than 2 */}
                {actions.length > 2 && (
                  <div className="relative">
                    <button
                      onClick={() =>
                        setOpenDropdown(
                          openDropdown === `mobile-${cardKeyExtractor(row)}`
                            ? null
                            : `mobile-${cardKeyExtractor(row)}`
                        )
                      }
                      className={`${cosmic.button.ghost} text-xs flex items-center space-x-1`}
                      title="More actions"
                    >
                      <FaEllipsisV />
                      <span>More</span>
                    </button>
                    {openDropdown === `mobile-${cardKeyExtractor(row)}` && (
                      <div className="absolute right-0 bottom-full mb-1 bg-gray-800 border border-gray-600 rounded-lg shadow-lg z-50 min-w-48">
                        {actions.slice(2).map((action, actionIndex) => {
                          if (action.condition && !action.condition(row)) {
                            return null;
                          }
                          return (
                            <button
                              key={actionIndex + 2}
                              onClick={() => {
                                action.onClick(row, index);
                                setOpenDropdown(null);
                              }}
                              className="w-full px-4 py-2 text-left text-sm text-white hover:bg-gray-700 flex items-center space-x-2 first:rounded-t-lg last:rounded-b-lg"
                            >
                              {action.icon}
                              <span>{action.label}</span>
                            </button>
                          );
                        })}
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );

  // Render pagination
  const renderPagination = () => {
    if (!pagination) return null;

    const { page, limit, total, onPageChange } = pagination;
    const totalPages = Math.ceil(total / limit);
    const startItem = (page - 1) * limit + 1;
    const endItem = Math.min(page * limit, total);

    return (
      <div className="flex items-center justify-between px-4 py-3 bg-white/5 border-t border-white/10 sm:px-6">
        <div className="flex-1 flex justify-between sm:hidden">
          <button
            onClick={() => onPageChange(page - 1)}
            disabled={page <= 1}
            className={`${cosmic.button.secondary} disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            Previous
          </button>
          <button
            onClick={() => onPageChange(page + 1)}
            disabled={page >= totalPages}
            className={`${cosmic.button.secondary} disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            Next
          </button>
        </div>
        <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
          <div>
            <p className="text-sm text-gray-300">
              Showing <span className="font-medium">{startItem}</span> to{" "}
              <span className="font-medium">{endItem}</span> of{" "}
              <span className="font-medium">{total}</span> results
            </p>
          </div>
          <div>
            <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
              <button
                onClick={() => onPageChange(page - 1)}
                disabled={page <= 1}
                className={`relative inline-flex items-center px-2 py-2 rounded-l-md ${cosmic.button.secondary} disabled:opacity-50 disabled:cursor-not-allowed`}
              >
                <FaChevronLeft className="h-3 w-3" />
              </button>

              {/* Page numbers */}
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                let pageNum;
                if (totalPages <= 5) {
                  pageNum = i + 1;
                } else if (page <= 3) {
                  pageNum = i + 1;
                } else if (page >= totalPages - 2) {
                  pageNum = totalPages - 4 + i;
                } else {
                  pageNum = page - 2 + i;
                }

                return (
                  <button
                    key={pageNum}
                    onClick={() => onPageChange(pageNum)}
                    className={`relative inline-flex items-center px-4 py-2 text-sm font-medium ${
                      page === pageNum
                        ? "bg-purple-600 text-white border-purple-600"
                        : `${cosmic.button.secondary} border-white/10`
                    }`}
                  >
                    {pageNum}
                  </button>
                );
              })}

              <button
                onClick={() => onPageChange(page + 1)}
                disabled={page >= totalPages}
                className={`relative inline-flex items-center px-2 py-2 rounded-r-md ${cosmic.button.secondary} disabled:opacity-50 disabled:cursor-not-allowed`}
              >
                <FaChevronRight className="h-3 w-3" />
              </button>
            </nav>
          </div>
        </div>
      </div>
    );
  };

  if (loading) {
    return (
      <div className={`${cosmic.cardElevated} ${className}`}>
        <div className="flex justify-center items-center h-64">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-500 mx-auto mb-4"></div>
            <p className="text-white">Loading...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={`${cosmic.cardElevated} ${className}`}>
      {/* Header with search and filters */}
      {(searchable || filterable) && (
        <div className="p-4 border-b border-white/10">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-3 sm:space-y-0">
            {searchable && (
              <div className="relative flex-1 max-w-lg">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <FaSearch className="h-4 w-4 text-gray-400" />
                </div>
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => handleSearchChange(e.target.value)}
                  className={`${cosmic.input} pl-10`}
                  placeholder="Search..."
                />
              </div>
            )}
            {filterable && (
              <button
                onClick={() => setShowFilters(!showFilters)}
                className={`${cosmic.button.secondary} flex items-center space-x-2`}
              >
                <FaFilter />
                <span>Filters</span>
              </button>
            )}
          </div>
        </div>
      )}

      {/* Data content */}
      <div className="overflow-x-hidden overflow-y-visible">
        {sortedData.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-400">{emptyMessage}</p>
          </div>
        ) : (
          <>
            {/* Desktop table view */}
            <div className={`hidden md:block`}>{renderTable()}</div>

            {/* Mobile card view */}
            <div className={`block md:hidden p-4`}>{renderCards()}</div>
          </>
        )}
      </div>

      {/* Pagination */}
      {renderPagination()}
    </div>
  );
}
