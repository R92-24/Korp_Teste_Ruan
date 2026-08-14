import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';

import { NotaService } from './nota';

describe('NotaService', () => {
  let service: NotaService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(NotaService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
