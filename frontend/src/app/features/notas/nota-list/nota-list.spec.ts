import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NotaList } from './nota-list';

describe('NotaList', () => {
  let component: NotaList;
  let fixture: ComponentFixture<NotaList>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NotaList],
    }).compileComponents();

    fixture = TestBed.createComponent(NotaList);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
