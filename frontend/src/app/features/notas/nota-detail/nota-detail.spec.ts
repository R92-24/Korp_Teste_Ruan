import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NotaDetail } from './nota-detail';

describe('NotaDetail', () => {
  let component: NotaDetail;
  let fixture: ComponentFixture<NotaDetail>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NotaDetail],
    }).compileComponents();

    fixture = TestBed.createComponent(NotaDetail);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
