function PageHead({ title, subtitle, children }) {
  return (
    <div className="st-pagehead">
      <div>
        <h1 className="st-pagehead__title">{title}</h1>
        <p className="st-pagehead__sub">{subtitle}</p>
      </div>
      <div className="st-pagehead__actions">{children}</div>
    </div>
  );
}

export default PageHead;
