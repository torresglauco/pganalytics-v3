import { Page, expect } from '@playwright/test';

/**
 * Page Object Model for Replication Topology Page
 * Handles navigation and interaction with replication topology visualization
 */
export class ReplicationTopologyPage {
  readonly page: Page;

  // Locators
  private readonly graphContainer = '.react-flow';
  private readonly nodeSelector = '.react-flow__node';
  private readonly edgeSelector = '.react-flow__edge';
  private readonly legendSidebar = '[data-testid="topology-legend"], .topology-legend, aside';
  private readonly refreshButton = 'button:has-text("Refresh")';
  private readonly loadingSpinner = '[data-testid="loading"], .spinner, .animate-spin';
  private readonly errorMessage = '.bg-red-50, [data-testid="error"], .alert-danger';
  private readonly emptyState = 'text=No Topology Data';
  private readonly pageTitle = 'h1:has-text("Replication Topology")';

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * Navigate to replication topology page for a specific collector
   */
  async goto(collectorId: string) {
    await this.page.goto(`/replication/topology/${collectorId}`);
    await this.expectPageLoaded();
  }

  /**
   * Wait for page to be fully loaded
   */
  async expectPageLoaded() {
    // Wait for either the page title or error/empty state
    await Promise.race([
      expect(this.page.locator(this.pageTitle)).toBeVisible({ timeout: 10000 }),
      expect(this.page.locator(this.errorMessage)).toBeVisible({ timeout: 10000 }),
      expect(this.page.locator(this.emptyState)).toBeVisible({ timeout: 10000 }),
    ]);

    // Wait for loading to complete if visible
    const spinner = this.page.locator(this.loadingSpinner).first();
    if (await spinner.isVisible({ timeout: 1000 }).catch(() => false)) {
      await spinner.waitFor({ state: 'hidden', timeout: 15000 });
    }
  }

  /**
   * Wait for topology graph to render
   */
  async expectGraphVisible() {
    const graph = this.page.locator(this.graphContainer);
    await expect(graph).toBeVisible({ timeout: 10000 });
  }

  /**
   * Get the number of visible nodes in the graph
   */
  async getNodeCount(): Promise<number> {
    await this.expectGraphVisible();
    return await this.page.locator(this.nodeSelector).count();
  }

  /**
   * Click on a specific node by index or id
   */
  async clickNode(nodeIdOrIndex: string | number) {
    await this.expectGraphVisible();

    let nodeLocator;
    if (typeof nodeIdOrIndex === 'number') {
      nodeLocator = this.page.locator(this.nodeSelector).nth(nodeIdOrIndex);
    } else {
      // Try to find node by ID in data attributes or content
      nodeLocator = this.page.locator(`${this.nodeSelector}[data-id="${nodeIdOrIndex}"]`).first();
      if (!(await nodeLocator.isVisible({ timeout: 1000 }).catch(() => false))) {
        // Fallback: find node containing the label text
        nodeLocator = this.page.locator(`${this.nodeSelector}:has-text("${nodeIdOrIndex}")`).first();
      }
    }

    await nodeLocator.click();
  }

  /**
   * Verify legend sidebar is visible
   */
  async expectLegendVisible() {
    const legend = this.page.locator(this.legendSidebar).first();
    await expect(legend).toBeVisible({ timeout: 5000 });
  }

  /**
   * Click the refresh button to reload topology
   */
  async refresh() {
    const button = this.page.locator(this.refreshButton);
    await expect(button).toBeEnabled({ timeout: 5000 });
    await button.click();

    // Wait for loading to start and complete
    const spinner = this.page.locator(this.loadingSpinner).first();
    if (await spinner.isVisible({ timeout: 1000 }).catch(() => false)) {
      await spinner.waitFor({ state: 'hidden', timeout: 15000 });
    }
  }

  /**
   * Check if error message is displayed
   */
  async hasError(): Promise<boolean> {
    const error = this.page.locator(this.errorMessage);
    return await error.isVisible({ timeout: 1000 }).catch(() => false);
  }

  /**
   * Get error message text if displayed
   */
  async getErrorMessage(): Promise<string | null> {
    const error = this.page.locator(this.errorMessage);
    if (await error.isVisible({ timeout: 1000 }).catch(() => false)) {
      return await error.textContent();
    }
    return null;
  }

  /**
   * Check if empty state is displayed (no topology data)
   */
  async hasEmptyState(): Promise<boolean> {
    const emptyState = this.page.locator(this.emptyState);
    return await emptyState.isVisible({ timeout: 1000 }).catch(() => false);
  }

  /**
   * Get all visible node labels
   */
  async getNodeLabels(): Promise<string[]> {
    await this.expectGraphVisible();
    const nodes = this.page.locator(this.nodeSelector);
    const count = await nodes.count();
    const labels: string[] = [];

    for (let i = 0; i < count; i++) {
      const text = await nodes.nth(i).textContent();
      if (text) {
        labels.push(text.trim());
      }
    }

    return labels;
  }

  /**
   * Get edge count in the graph
   */
  async getEdgeCount(): Promise<number> {
    await this.expectGraphVisible();
    return await this.page.locator(this.edgeSelector).count();
  }
}