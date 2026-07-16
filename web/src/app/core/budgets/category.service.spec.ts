import { provideHttpClient } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { CategoryService } from './category.service';

describe('CategoryService', () => {
  let service: CategoryService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(CategoryService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    TestBed.resetTestingModule();
  });

  it('deletes an empty category without a reassignment query', async () => {
    const deletion = service.remove(12, 34);

    const deleteRequest = http.expectOne('/api/categories/34');
    expect(deleteRequest.request.method).toBe('DELETE');
    expect(deleteRequest.request.params.has('reassignToCategoryId')).toBe(false);
    deleteRequest.flush(null, { status: 204, statusText: 'No Content' });
    await Promise.resolve();

    const loadRequest = http.expectOne('/api/budgets/12/categories');
    expect(loadRequest.request.method).toBe('GET');
    loadRequest.flush([]);

    await deletion;
    expect(service.categories()).toEqual([]);
  });

  it('sends the selected reassignment category before reloading', async () => {
    const deletion = service.remove(12, 34, 35);

    const deleteRequest = http.expectOne(
      (request) =>
        request.url === '/api/categories/34' && request.params.get('reassignToCategoryId') === '35',
    );
    expect(deleteRequest.request.method).toBe('DELETE');
    deleteRequest.flush(null, { status: 204, statusText: 'No Content' });
    await Promise.resolve();

    const loadRequest = http.expectOne('/api/budgets/12/categories');
    loadRequest.flush([]);

    await deletion;
  });
});
