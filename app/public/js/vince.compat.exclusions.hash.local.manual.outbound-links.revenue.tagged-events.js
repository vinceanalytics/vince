!function(){"use strict";var e,t,u=window.location,o=window.document,l=o.getElementById("plausible"),s=l.getAttribute("data-api")||(e=(e=l).src.split("/"),t=e[0],e=e[2],t+"//"+e+"/api/event");function p(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function r(e,t){try{if("true"===window.localStorage.plausible_ignore)return p("localStorage flag",t)}catch(e){}var r=l&&l.getAttribute("data-include"),a=l&&l.getAttribute("data-exclude");if("pageview"===e){r=!r||r.split(",").some(n),a=a&&a.split(",").some(n);if(!r||a)return p("exclusion rule",t)}function n(e){var t=u.pathname;return(t+=u.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var r={},i=(r.n=e,r.u=t&&t.u?t.u:u.href,r.d=l.getAttribute("data-domain"),r.r=o.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),t&&t.revenue&&(r.$=t.revenue),r.h=1,new XMLHttpRequest);i.open("POST",s,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(r)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var a=window.plausible&&window.plausible.q||[];window.plausible=r;for(var n=0;n<a.length;n++)r.apply(this,a[n]);function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var f=1;function i(e){var t,r;if("auxclick"!==e.type||e.button===f)return(t=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],!function e(t,r){if(!t||m<r)return!1;if(b(t))return!0;return e(t.parentNode,r+1)}(t,0)&&(r=t)&&r.href&&r.host&&r.host!==u.host?v(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):void 0}function v(e,t,r){var a,n=!1;function i(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((a={props:r.props}).revenue=r.revenue,plausible(r.name,a)):((a={props:r.props,callback:i}).revenue=r.revenue,plausible(r.name,a),setTimeout(i,5e3),e.preventDefault())}function d(e){var e=b(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},r=e&&e.classList;if(r)for(var a=0;a<r.length;a++){var n,i,u=r.item(a),o=u.match(/plausible-event-(.+)(=|--)(.+)/),o=(o&&(n=o[1],i=o[3].replace(/\+/g," "),"name"==n.toLowerCase()?t.name=i:t.props[n]=i),u.match(/plausible-revenue-(.+)(=|--)(.+)/));o&&(n=o[1],i=o[3],t.revenue[n]=i)}return t}o.addEventListener("click",i),o.addEventListener("auxclick",i);var m=3;function g(e){if("auxclick"!==e.type||e.button===f){for(var t,r,a,n,i=e.target,u=0;u<=m&&i;u++){if((a=i)&&a.tagName&&"form"===a.tagName.toLowerCase())return;c(i)&&(t=i),b(i)&&(r=i),i=i.parentNode}r&&(n=d(r),t?(n.props.url=t.href,v(e,t,n)):((e={}).props=n.props,e.revenue=n.revenue,plausible(n.name,e)))}}function b(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,r=e.target,a=d(r);function n(){t||(t=!0,r.submit())}a.name&&(e.preventDefault(),t=!1,setTimeout(n,5e3),(e={props:a.props,callback:n}).revenue=a.revenue,plausible(a.name,e))}),o.addEventListener("click",g),o.addEventListener("auxclick",g)}();