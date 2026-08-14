import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';

import { NotaDetail } from './nota-detail';

describe('NotaDetail', () => {
  let component: NotaDetail;
  let fixture: ComponentFixture<NotaDetail>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NotaDetail],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(NotaDetail);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
