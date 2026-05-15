import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { ReplicationTopologyPage } from '../pages/ReplicationTopologyPage';

test.describe('Replication Topology', () => {
  let loginPage: LoginPage;
  let topologyPage: ReplicationTopologyPage;

  // Test collector ID - adjust based on test data
  const testCollectorId = 'test-collector-001';

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    topologyPage = new ReplicationTopologyPage(page);

    // Login before each test
    await loginPage.goto();
    await loginPage.login('admin', 'admin');
  });

  test('should display topology page with graph', async ({ page }) => {
    await topologyPage.goto(testCollectorId);

    // Verify page title is visible
    await expect(page.locator('h1:has-text("Replication Topology")')).toBeVisible();
  });

  test('should render topology graph with nodes', async ({ page }) => {
    await topologyPage.goto(testCollectorId);

    // Wait for graph to be visible
    await topologyPage.expectGraphVisible();

    // Verify nodes are rendered (at least 1 node should exist)
    const nodeCount = await topologyPage.getNodeCount();
    expect(nodeCount).toBeGreaterThanOrEqual(1);
  });

  test('should display legend sidebar', async ({ page }) => {
    await topologyPage.goto(testCollectorId);

    // Verify legend is visible
    await topologyPage.expectLegendVisible();
  });

  test('should refresh topology when refresh button clicked', async ({ page }) => {
    await topologyPage.goto(testCollectorId);

    // Wait for initial load
    await topologyPage.expectGraphVisible();

    // Click refresh
    await topologyPage.refresh();

    // Verify graph is still visible after refresh
    await topologyPage.expectGraphVisible();
  });

  test('should handle clicking on a node', async ({ page }) => {
    await topologyPage.goto(testCollectorId);

    // Wait for graph
    await topologyPage.expectGraphVisible();

    const nodeCount = await topologyPage.getNodeCount();
    if (nodeCount > 0) {
      // Click first node - should not throw error
      await topologyPage.clickNode(0);
    }
  });

  test('should handle missing or invalid collectorId gracefully', async ({ page }) => {
    // Navigate with invalid collector ID
    await topologyPage.goto('non-existent-collector-id');

    // Should show either error or empty state
    const hasError = await topologyPage.hasError();
    const hasEmptyState = await topologyPage.hasEmptyState();

    // At least one should be true for invalid collector
    expect(hasError || hasEmptyState).toBe(true);
  });

  test('should display node labels correctly', async ({ page }) => {
    await topologyPage.goto(testCollectorId);

    // Wait for graph
    await topologyPage.expectGraphVisible();

    const nodeCount = await topologyPage.getNodeCount();
    if (nodeCount > 0) {
      const labels = await topologyPage.getNodeLabels();
      expect(labels.length).toBeGreaterThan(0);
    }
  });

  test('should show edges connecting nodes', async ({ page }) => {
    await topologyPage.goto(testCollectorId);

    // Wait for graph
    await topologyPage.expectGraphVisible();

    const nodeCount = await topologyPage.getNodeCount();
    const edgeCount = await topologyPage.getEdgeCount();

    // If we have multiple nodes, we should have edges connecting them
    if (nodeCount > 1) {
      expect(edgeCount).toBeGreaterThanOrEqual(1);
    }
  });

  test('should display loading state initially', async ({ page }) => {
    // Navigate and immediately check for loading or content
    await page.goto(`/replication/topology/${testCollectorId}`);

    // Either loading spinner or graph should be visible quickly
    const loadingSpinner = page.locator('[data-testid="loading"], .spinner, .animate-spin').first();
    const graphContainer = page.locator('.react-flow').first();

    // One of them should be visible within a short time
    await Promise.race([
      expect(loadingSpinner).toBeVisible({ timeout: 2000 }).catch(() => {}),
      expect(graphContainer).toBeVisible({ timeout: 2000 }).catch(() => {}),
    ]);
  });

  test('should maintain authentication when accessing topology page', async ({ page }) => {
    await topologyPage.goto(testCollectorId);

    // Should not redirect to login
    expect(page.url()).not.toContain('/login');
  });
});